//! Конфигурация вычислительного сервиса.
//!
//! Все параметры читаются из переменных окружения с разумными значениями
//! по умолчанию. Конфигурация загружается один раз при старте процесса
//! через [`Config::from_env`].

use std::env;

/// Полная конфигурация вычислительного сервиса.
///
/// Каждое поле соответствует одной переменной окружения. Значения по
/// умолчанию подобраны так, чтобы сервис запускался «из коробки» без
/// дополнительной настройки.
#[derive(Debug, Clone)]
pub struct Config {
    /// URL подключения к Redis (env `REDIS_URL`).
    pub redis_url: String,
    /// Пространство имён для ключей Redis (env `PII_NS`).
    pub ns: String,
    /// Время жизни записей соответствия в секундах (env `PII_CORR_TTL_SEC`).
    pub corr_ttl_sec: u64,
    /// Максимальный размер in-memory кэша конфигов (env `PII_CACHE_MAX`).
    pub cache_max: usize,
    /// Максимальное число одновременных обработок запросов (env `PII_MAX_CONCURRENT`).
    pub max_concurrent: usize,
    /// Время ожидания слота семафора в секундах (env `PII_SEM_WAIT_SEC`).
    pub sem_wait_sec: f64,
    /// Размер чанка для NER-детекции в символах (env `PII_NER_CHUNK_CHARS`).
    pub ner_chunk_chars: usize,
    /// Перекрытие соседних чанков в символах (env `PII_NER_OVERLAP`).
    pub ner_overlap: usize,
    /// Размер контекстного окна вокруг спана в символах (env `PII_CONTEXT_WINDOW`).
    pub context_window: usize,
    /// Ключ FPE-шифрования (env `PII_FPE_KEY`).
    pub fpe_key: String,
    /// Хост для HTTP-сервера (env `PII_APP_HOST`).
    pub app_host: String,
    /// Порт для HTTP-сервера (env `PII_APP_PORT`).
    pub app_port: u16,
}

impl Config {
    /// Загружает конфигурацию из переменных окружения.
    ///
    /// Для каждого параметра используется переменная окружения, указанная
    /// в документации поля; если переменная не задана или не может быть
    /// разобрана, применяется значение по умолчанию.
    pub fn from_env() -> Self {
        Config {
            redis_url: env::var("REDIS_URL").unwrap_or_else(|_| "redis://localhost:6379/0".into()),
            ns: env::var("PII_NS").unwrap_or_else(|_| "pii".into()),
            corr_ttl_sec: env::var("PII_CORR_TTL_SEC")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(3600),
            cache_max: env::var("PII_CACHE_MAX")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(100_000),
            max_concurrent: env::var("PII_MAX_CONCURRENT")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or_else(num_cpus::get),
            sem_wait_sec: env::var("PII_SEM_WAIT_SEC")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(0.3),
            ner_chunk_chars: env::var("PII_NER_CHUNK_CHARS")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(4000),
            ner_overlap: env::var("PII_NER_OVERLAP")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(200),
            context_window: env::var("PII_CONTEXT_WINDOW")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(200),
            fpe_key: env::var("PII_FPE_KEY").unwrap_or_default(),
            app_host: env::var("PII_APP_HOST").unwrap_or_else(|_| "0.0.0.0".into()),
            app_port: env::var("PII_APP_PORT")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(8080),
        }
    }
}