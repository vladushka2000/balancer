//! Репозиторий конфигураций систем-потребителей.
//!
//! Хранит [`SystemConfig`] в Redis и кэширует их in-memory на короткое
//! время (2 секунды), чтобы снизить нагрузку на Redis при частых чтениях.

use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use redis::aio::ConnectionManager;
use redis::AsyncCommands;

use crate::keys::{config_epoch_key, system_key, systems_set};
use crate::models::SystemConfig;

/// Время жизни in-memory кэша конфигов.
const CACHE_TTL: Duration = Duration::from_secs(2);

/// Внутреннее состояние кэша конфигов.
struct SystemCache {
    /// Карта «system_id → (конфиг, время записи)».
    entries: HashMap<String, (SystemConfig, Instant)>,
}

impl SystemCache {
    /// Создаёт пустой кэш.
    fn new() -> Self {
        SystemCache {
            entries: HashMap::new(),
        }
    }

    /// Возвращает конфиг, если он есть в кэше и ещё не истёк.
    fn get(&self, system_id: &str) -> Option<SystemConfig> {
        let (config, written_at) = self.entries.get(system_id)?;
        if written_at.elapsed() > CACHE_TTL {
            return None;
        }
        Some(config.clone())
    }

    /// Кладёт конфиг в кэш с текущей временной меткой.
    fn put(&mut self, system_id: &str, config: SystemConfig) {
        self.entries
            .insert(system_id.to_string(), (config, Instant::now()));
    }

    /// Очищает кэш (вызывается при сохранении конфига).
    fn invalidate(&mut self) {
        self.entries.clear();
    }
}

/// Репозиторий конфигураций систем.
#[derive(Clone)]
pub struct Repo {
    /// Менеджер подключения к Redis.
    conn: ConnectionManager,
    /// Пространство имён ключей.
    ns: String,
    /// In-memory кэш конфигов.
    cache: Arc<Mutex<SystemCache>>,
}

impl Repo {
    /// Создаёт репозиторий поверх подключения к Redis.
    pub fn new(conn: ConnectionManager, ns: String) -> Self {
        Repo {
            conn,
            ns,
            cache: Arc::new(Mutex::new(SystemCache::new())),
        }
    }

    /// Сохраняет конфиг системы в Redis и инвалидирует кэш.
    ///
    /// Запись выполняется транзакционно: конфиг кладётся по ключу системы,
    /// а идентификатор добавляется во множество всех систем.
    pub async fn save_system(&self, config: &SystemConfig) -> redis::RedisResult<()> {
        let mut conn = self.conn.clone();
        let key = system_key(&self.ns, &config.system_id);
        let set = systems_set(&self.ns);
        let json = serde_json::to_string(config).map_err(serde_err)?;
        redis::pipe()
            .set(&key, json)
            .sadd(&set, &config.system_id)
            .query_async::<_, ()>(&mut conn)
            .await?;
        self.cache.lock().unwrap().invalidate();
        Ok(())
    }

    /// Возвращает конфиг системы, используя in-memory кэш.
    ///
    /// Сначала проверяется кэш; при промахе конфиг читается из Redis и
    /// кладётся в кэш.
    pub async fn get_system(&self, system_id: &str) -> redis::RedisResult<Option<SystemConfig>> {
        if let Some(config) = self.cache.lock().unwrap().get(system_id) {
            return Ok(Some(config));
        }
        let mut conn = self.conn.clone();
        let key = system_key(&self.ns, system_id);
        let raw: Option<String> = conn.get(&key).await?;
        let config = match raw {
            Some(json) => {
                let config: SystemConfig = serde_json::from_str(&json).map_err(serde_err)?;
                self.cache.lock().unwrap().put(system_id, config.clone());
                Some(config)
            }
            None => None,
        };
        Ok(config)
    }

    /// Возвращает список всех зарегистрированных систем.
    pub async fn list_systems(&self) -> redis::RedisResult<Vec<SystemConfig>> {
        let mut conn = self.conn.clone();
        let set = systems_set(&self.ns);
        let ids: Vec<String> = conn.smembers(&set).await?;
        let mut systems = Vec::with_capacity(ids.len());
        for id in ids {
            if let Some(config) = self.get_system(&id).await? {
                systems.push(config);
            }
        }
        Ok(systems)
    }

    /// Увеличивает эпоху конфигурации (сигнал о смене конфигов).
    pub async fn bump_config_epoch(&self) -> redis::RedisResult<()> {
        let mut conn = self.conn.clone();
        let key = config_epoch_key(&self.ns);
        conn.incr::<_, _, ()>(&key, 1).await?;
        Ok(())
    }

    /// Возвращает текущее значение эпохи конфигурации.
    pub async fn get_control_epoch(&self) -> redis::RedisResult<i64> {
        let mut conn = self.conn.clone();
        let key = config_epoch_key(&self.ns);
        let value: i64 = conn.get(&key).await.unwrap_or(0);
        Ok(value)
    }
}

/// Преобразует ошибку сериализации в ошибку Redis.
fn serde_err(e: serde_json::Error) -> redis::RedisError {
    redis::RedisError::from((redis::ErrorKind::ResponseError, "serde", e.to_string()))
}
