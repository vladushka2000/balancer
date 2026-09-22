//! NER-детекция ПД: regex+dict, чанки, spawn_blocking.

use std::collections::HashSet;

use crate::models::Span;

/// Предзагруженные regex-паттерны NER.
pub struct NERPatterns;

/// NER-детектор ФИО, адресов и органов.
pub struct NERDetector {
    /// Предзагруженные regex.
    patterns: NERPatterns,
    /// Словарь известных имён (famous.txt, casefold).
    famous: HashSet<String>,
    /// Размер чанка в символах.
    chunk_chars: usize,
    /// Перекрытие соседних чанков в символах.
    overlap: usize,
}

impl NERDetector {
    /// Создаёт NER-детектор с заданными параметрами чанкования.
    pub fn new(chunk_chars: usize, overlap: usize) -> Self {
        todo!("реализация создания NER-детектора")
    }

    /// Загружает regex и словари.
    pub fn preload(&mut self) {
        todo!("реализация предзагрузки regex и словарей")
    }

    /// Ищет ПД в тексте, обрабатывая чанки в `spawn_blocking`.
    pub async fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация NER-детекции")
    }
}
