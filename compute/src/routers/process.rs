//! Роутер обработки запросов (`POST /process`).

use std::sync::Arc;

use axum::extract::State;
use axum::Json;
use axum::response::IntoResponse;

use crate::app_state::AppState;
use crate::models::ProcessRequest;

/// Обработчик `POST /process`.
///
/// Идемпотентность → семафор → конвейер → хранилище → статистика.
pub async fn process_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ProcessRequest>,
) -> impl IntoResponse {
    todo!("реализация обработчика /process")
}