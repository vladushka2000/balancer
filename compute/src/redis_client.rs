//! Создание подключения к Redis.
//!
//! Используется [`ConnectionManager`] для автоматического переподключения
//! при обрыве соединения.

use redis::aio::ConnectionManager;

/// Создаёт менеджер подключения к Redis по URL.
///
/// # Аргументы
///
/// * `url` — строка подключения, например `redis://localhost:6379/0`.
///
/// # Ошибки
///
/// Возвращает ошибку Redis, если подключение не удалось установить.
pub async fn make_redis(url: &str) -> redis::RedisResult<ConnectionManager> {
    let client = redis::Client::open(url)?;
    ConnectionManager::new(client).await
}