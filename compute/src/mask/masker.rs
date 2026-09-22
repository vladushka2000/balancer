//! Маскер — применение правил маскирования к спанам.

use crate::mask::fpe::FPEMasker;
use crate::models::Span;

/// Маскер, применяющий правила к спанам справа налево.
pub struct Masker {
    /// Режим маскирования (`partial`, `fpe`, `redact`).
    mode: String,
    /// Опциональный FPE-маскер.
    fpe: Option<FPEMasker>,
}

impl Masker {
    /// Создаёт маскер с заданным режимом и ключом FPE.
    pub fn new(mode: &str, fpe_key: &str) -> Self {
        todo!("реализация создания маскера")
    }

    /// Применяет маскирование к тексту по спанам (справа налево).
    pub fn apply(&self, text: &str, spans: &[Span]) -> String {
        todo!("реализация применения маскирования")
    }
}
