//! Точка входа вычислительного сервиса PII-модуля.
//!
//! Загружает конфигурацию, инициализирует логирование и поднимает
//! HTTP-сервер на базе axum с эндпоинтом `/health`.

use axum::{routing::get, Router};

use compute::config::Config;

/// Обработчик эндпоинта `/health`.
///
/// Возвращает `200 OK` с телом `ok`, подтверждая, что сервис жив.
async fn health() -> &'static str {
    "ok"
}

/// Строит HTTP-роутер приложения.
fn build_router() -> Router {
    Router::new().route("/health", get(health))
}

/// Точка входа сервиса.
#[tokio::main]
async fn main() {
    let config = Config::from_env();

    let addr = format!("{}:{}", config.app_host, config.app_port);
    let listener = tokio::net::TcpListener::bind(&addr)
        .await
        .unwrap_or_else(|e| panic!("не удалось привязаться к {addr}: {e}"));

    tracing::info!("compute-сервис запущен на {addr}");

    axum::serve(listener, build_router())
        .await
        .unwrap_or_else(|e| panic!("ошибка HTTP-сервера: {e}"));
}
