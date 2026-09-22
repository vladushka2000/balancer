//! Тесты загрузки конфигурации из переменных окружения.

use std::sync::Mutex;

use compute::config::Config;

/// Сериализует доступ к переменным окружения между тестами.
static ENV_LOCK: Mutex<()> = Mutex::new(());

/// Очищает все переменные окружения, влияющие на конфигурацию.
fn clear_env() {
    for var in [
        "REDIS_URL",
        "PII_NS",
        "PII_CORR_TTL_SEC",
        "PII_CACHE_MAX",
        "PII_MAX_CONCURRENT",
        "PII_SEM_WAIT_SEC",
        "PII_NER_CHUNK_CHARS",
        "PII_NER_OVERLAP",
        "PII_CONTEXT_WINDOW",
        "PII_FPE_KEY",
        "PII_APP_HOST",
        "PII_APP_PORT",
    ] {
        std::env::remove_var(var);
    }
}

/// Проверяет значения по умолчанию при пустых переменных окружения.
#[test]
fn test_config_defaults() {
    let _guard = ENV_LOCK.lock().unwrap();
    clear_env();

    let config = Config::from_env();
    assert_eq!(config.redis_url, "redis://localhost:6379/0");
    assert_eq!(config.ns, "pii");
    assert_eq!(config.corr_ttl_sec, 3600);
    assert_eq!(config.cache_max, 100_000);
    assert_eq!(config.max_concurrent, num_cpus::get());
    assert_eq!(config.sem_wait_sec, 0.3);
    assert_eq!(config.ner_chunk_chars, 4000);
    assert_eq!(config.ner_overlap, 200);
    assert_eq!(config.context_window, 200);
    assert_eq!(config.fpe_key, "");
    assert_eq!(config.app_host, "0.0.0.0");
    assert_eq!(config.app_port, 8080);
}

/// Проверяет, что значения подхватываются из переменных окружения.
#[test]
fn test_config_from_env() {
    let _guard = ENV_LOCK.lock().unwrap();
    clear_env();

    std::env::set_var("REDIS_URL", "redis://example:6379/2");
    std::env::set_var("PII_NS", "prod");
    std::env::set_var("PII_CORR_TTL_SEC", "7200");
    std::env::set_var("PII_CACHE_MAX", "5000");
    std::env::set_var("PII_MAX_CONCURRENT", "8");
    std::env::set_var("PII_SEM_WAIT_SEC", "1.5");
    std::env::set_var("PII_NER_CHUNK_CHARS", "8000");
    std::env::set_var("PII_NER_OVERLAP", "400");
    std::env::set_var("PII_CONTEXT_WINDOW", "300");
    std::env::set_var("PII_FPE_KEY", "secret-key");
    std::env::set_var("PII_APP_HOST", "127.0.0.1");
    std::env::set_var("PII_APP_PORT", "9090");

    let config = Config::from_env();
    assert_eq!(config.redis_url, "redis://example:6379/2");
    assert_eq!(config.ns, "prod");
    assert_eq!(config.corr_ttl_sec, 7200);
    assert_eq!(config.cache_max, 5000);
    assert_eq!(config.max_concurrent, 8);
    assert_eq!(config.sem_wait_sec, 1.5);
    assert_eq!(config.ner_chunk_chars, 8000);
    assert_eq!(config.ner_overlap, 400);
    assert_eq!(config.context_window, 300);
    assert_eq!(config.fpe_key, "secret-key");
    assert_eq!(config.app_host, "127.0.0.1");
    assert_eq!(config.app_port, 9090);
}