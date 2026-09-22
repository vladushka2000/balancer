//! Административный роутер (`/systems`, `/stats`, `/health`, `/clear`).

use std::sync::Arc;

use axum::extract::State;
use axum::response::IntoResponse;

use crate::app_state::AppState;

/// Обработчик `POST /systems` — сохранение конфигурации системы.
pub async fn save_system(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    todo!("реализация сохранения системы")
}

/// Обработчик `GET /systems` — список систем.
pub async fn list_systems(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    todo!("реализация списка систем")
}

/// Обработчик `GET /stats` — статистика.
pub async fn get_stats(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    todo!("реализация статистики")
}

/// Обработчик `GET /health` — проверка живости.
pub async fn health() -> impl IntoResponse {
    todo!("реализация проверки живости")
}

/// Обработчик `POST /clear` — очистка состояния.
pub async fn clear(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    todo!("реализация очистки состояния")
}