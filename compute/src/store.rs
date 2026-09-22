//! Хранилище соответствий «оригинал ↔ маска».
//!
//! Обеспечивает идемпотентное демаскирование: по идентификатору полезной
//! нагрузки хранится запись соответствия, позволяющая восстановить
//! оригинал из маски и наоборот. Используется LRU-кэш с TTL поверх Redis.

use std::num::NonZeroUsize;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use lru::LruCache;
use redis::aio::ConnectionManager;
use redis::AsyncCommands;

use crate::keys::corr_key;
use crate::models::CorrRecord;

/// Тип результата операций хранилища.
pub type Result<T> = std::result::Result<T, redis::RedisError>;

/// Внутреннее состояние LRU-кэша записей соответствия.
struct CacheInner {
    /// LRU-кэш «payload_id → (запись, время записи)».
    entries: LruCache<String, (CorrRecord, Instant)>,
}

/// Хранилище соответствий с in-memory LRU+TTL кэшем.
#[derive(Clone)]
pub struct CorrespondenceStore {
    /// Менеджер подключения к Redis.
    conn: ConnectionManager,
    /// Пространство имён ключей.
    ns: String,
    /// Время жизни записей в секундах.
    ttl_sec: u64,
    /// In-memory LRU-кэш.
    cache: Arc<Mutex<CacheInner>>,
}

impl CorrespondenceStore {
    /// Создаёт хранилище поверх подключения к Redis.
    ///
    /// # Аргументы
    ///
    /// * `conn` — менеджер подключения к Redis.
    /// * `ns` — пространство имён ключей.
    /// * `ttl_sec` — время жизни записей в секундах.
    /// * `cache_max` — максимальный размер in-memory кэша.
    pub fn new(conn: ConnectionManager, ns: String, ttl_sec: u64, cache_max: usize) -> Self {
        CorrespondenceStore {
            conn,
            ns,
            ttl_sec,
            cache: Arc::new(Mutex::new(CacheInner {
                entries: LruCache::new(NonZeroUsize::new(cache_max).unwrap_or(NonZeroUsize::MIN)),
            })),
        }
    }

    /// Возвращает запись соответствия по идентификатору полезной нагрузки.
    ///
    /// Сначала проверяется in-memory кэш; при промахе запись читается из
    /// Redis, парсится из JSON и кладётся в кэш.
    pub async fn get(&self, payload_id: &str) -> Result<Option<CorrRecord>> {
        {
            let mut cache = self.cache.lock().unwrap();
            if let Some((record, written_at)) = cache.entries.get(payload_id) {
                if written_at.elapsed() <= Duration::from_secs(self.ttl_sec) {
                    return Ok(Some(record.clone()));
                }
            }
        }

        let mut conn = self.conn.clone();
        let key = corr_key(&self.ns, payload_id);
        let raw: Option<String> = conn.get(&key).await?;
        let record = match raw {
            Some(json) => {
                let record: CorrRecord = serde_json::from_str(&json).map_err(serde_err)?;
                let mut cache = self.cache.lock().unwrap();
                cache
                    .entries
                    .put(payload_id.to_string(), (record.clone(), Instant::now()));
                Some(record)
            }
            None => None,
        };
        Ok(record)
    }

    /// Сохраняет запись соответствия (write-through: кэш + Redis с TTL).
    pub async fn put(
        &self,
        payload_id: &str,
        original: &str,
        mask: &str,
        types: Vec<String>,
    ) -> Result<()> {
        let record = CorrRecord {
            original: original.to_string(),
            mask: mask.to_string(),
            types,
            created_ts: now_ts(),
        };

        {
            let mut cache = self.cache.lock().unwrap();
            cache
                .entries
                .put(payload_id.to_string(), (record.clone(), Instant::now()));
        }

        let mut conn = self.conn.clone();
        let key = corr_key(&self.ns, payload_id);
        let json = serde_json::to_string(&record).map_err(serde_err)?;
        conn.set_ex::<_, _, ()>(&key, json, self.ttl_sec).await?;
        Ok(())
    }

    /// Идемпотентный поиск соответствия по полезной нагрузке.
    ///
    /// Если запись найдена и `payload` совпадает с оригиналом — возвращает
    /// маску; если `payload` совпадает с маской — возвращает оригинал;
    /// иначе возвращает `None`.
    pub async fn lookup(
        &self,
        payload_id: &str,
        payload: &str,
    ) -> Result<Option<(String, String)>> {
        let record = match self.get(payload_id).await? {
            Some(record) => record,
            None => return Ok(None),
        };
        if payload == record.original {
            Ok(Some((record.mask, record.original)))
        } else if payload == record.mask {
            Ok(Some((record.original, record.mask)))
        } else {
            Ok(None)
        }
    }
}

/// Возвращает текущую временную метку Unix-времени в секундах.
fn now_ts() -> f64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs_f64())
        .unwrap_or(0.0)
}

/// Преобразует ошибку сериализации в ошибку Redis.
fn serde_err(e: serde_json::Error) -> redis::RedisError {
    redis::RedisError::from((redis::ErrorKind::ResponseError, "serde", e.to_string()))
}