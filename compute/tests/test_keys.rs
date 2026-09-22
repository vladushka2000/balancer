//! Тесты хелперов построения ключей Redis.

use compute::keys::{config_epoch_key, corr_key, system_key, systems_set};

/// Проверяет формат ключа записи соответствия.
#[test]
fn test_corr_key() {
    assert_eq!(corr_key("pii", "id1"), "pii:corr:id1");
}

/// Проверяет формат ключа конфигурации системы.
#[test]
fn test_system_key() {
    assert_eq!(system_key("pii", "s1"), "pii:system:s1");
}

/// Проверяет формат ключа множества систем.
#[test]
fn test_systems_set() {
    assert_eq!(systems_set("pii"), "pii:systems");
}

/// Проверяет формат ключа эпохи конфигурации.
#[test]
fn test_config_epoch_key() {
    assert_eq!(config_epoch_key("pii"), "pii:control:config_epoch");
}