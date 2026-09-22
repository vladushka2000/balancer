//! Хелперы для построения ключей Redis.
//!
//! Все ключи строятся по единому шаблону `{ns}:{domain}:{id}`, что
//! позволяет изолировать данные разных окружений через пространство имён.

/// Ключ записи соответствия «оригинал ↔ маска» для полезной нагрузки.
///
/// Формат: `{ns}:corr:{payload_id}`.
pub fn corr_key(ns: &str, payload_id: &str) -> String {
    format!("{ns}:corr:{payload_id}")
}

/// Ключ конфигурации системы-потребителя.
///
/// Формат: `{ns}:system:{system_id}`.
pub fn system_key(ns: &str, system_id: &str) -> String {
    format!("{ns}:system:{system_id}")
}

/// Ключ множества (set) всех зарегистрированных систем.
///
/// Формат: `{ns}:systems`.
pub fn systems_set(ns: &str) -> String {
    format!("{ns}:systems")
}

/// Ключ эпохи конфигурации (монотонно растущий счётчик изменений).
///
/// Формат: `{ns}:control:config_epoch`.
pub fn config_epoch_key(ns: &str) -> String {
    format!("{ns}:control:config_epoch")
}