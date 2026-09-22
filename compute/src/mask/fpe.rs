//! Формат-сохраняющее шифрование (FPE).

use crate::models::Span;

/// FPE-маскер: цифры → цифры той же длины, буквы → буквы.
pub struct FPEMasker {
    /// Ключ шифрования.
    key: Vec<u8>,
}

impl FPEMasker {
    /// Создаёт FPE-маскер с заданным ключом.
    pub fn new(key: &str) -> Self {
        todo!("реализация создания FPE-маскера")
    }

    /// Маскирует фрагмент текста.
    pub fn mask(&self, text: &str, span: &Span) -> String {
        todo!("реализация FPE-маскирования")
    }

    /// Восстанавливает оригинал из маски.
    pub fn unmask(&self, masked: &str, span: &Span) -> String {
        todo!("реализация FPE-демаскирования")
    }
}
