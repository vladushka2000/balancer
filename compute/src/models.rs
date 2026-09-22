//! Модели данных — замороженный контракт между API и compute-сервисом.
//!
//! Поля и типы этих структур менять нельзя без явного решения: они
//! используются для сериализации/десериализации запросов и ответов.

use serde::{Deserialize, Serialize};

/// Запрос на обработку полезной нагрузки.
///
/// Приходит от API-шлюза и содержит текст запроса к LLM вместе с
/// идентификатором полезной нагрузки для идемпотентности.
#[derive(Serialize, Deserialize)]
pub struct ProcessRequest {
    /// Текст полезной нагрузки (запрос к LLM).
    pub payload: String,
    /// Идентификатор полезной нагрузки.
    pub payload_id: String,
}

/// Ответ на обработку полезной нагрузки.
#[derive(Serialize, Deserialize)]
pub struct ProcessResponse {
    /// Результат обработки (маскированный или демаскированный текст).
    pub result: String,
}

/// Спан — найденный фрагмент текста, содержащий ПД.
#[derive(Clone, Serialize, Deserialize)]
pub struct Span {
    /// Начало спана в символах (включительно).
    pub start: usize,
    /// Конец спана в символах (исключительно).
    pub end: usize,
    /// Тип ПД (например, `phone`, `email`, `name`).
    #[serde(rename = "type")]
    pub type_: String,
    /// Уверенность детектора в диапазоне [0, 1].
    pub confidence: f64,
    /// Источник спана (например, `structural`, `ner`, `context`).
    pub source: String,
}

/// Запись соответствия «оригинал ↔ маска» для демаскирования.
#[derive(Clone, Serialize, Deserialize)]
pub struct CorrRecord {
    /// Оригинальный фрагмент текста.
    pub original: String,
    /// Маскированный фрагмент текста.
    pub mask: String,
    /// Типы ПД, найденные в оригинале.
    pub types: Vec<String>,
    /// Временная метка создания записи (Unix-время, секунды).
    pub created_ts: f64,
}

/// Конфигурация системы-потребителя (правила маскирования).
#[derive(Clone, Serialize, Deserialize)]
pub struct SystemConfig {
    /// Идентификатор системы.
    pub system_id: String,
    /// Включена ли обработка для системы.
    pub enabled: bool,
    /// Список типов ПД для обработки (разделитель — запятая).
    pub types: String,
    /// Режим маскирования (`partial`, `fpe`, `synthetic`, `redact`).
    pub mask_mode: String,
    /// Разрешено ли демаскирование.
    pub demask_enabled: bool,
}
