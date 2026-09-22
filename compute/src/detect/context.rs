//! Контекстное правило принадлежности ПД.

use std::collections::HashSet;

use crate::models::Span;

/// Контекстное правило фильтрации NER-спанов.
pub struct ContextRule {
    /// Размер контекстного окна вокруг спана в символах.
    window: usize,
    /// Словарь известных имён (famous.txt).
    famous: HashSet<String>,
    /// Словарь организаций и адресов (org_addresses.txt).
    org_addresses: HashSet<String>,
    /// Словарь маркеров (markers.txt).
    markers: HashSet<String>,
}

impl ContextRule {
    /// Создаёт контекстное правило и загружает словари.
    pub fn new(window: usize) -> Self {
        todo!("реализация создания контекстного правила")
    }

    /// Фильтрует NER-спаны по словарям и контекстному окну.
    pub fn filter(&self, spans: Vec<Span>, text: &str) -> Vec<Span> {
        todo!("реализация фильтрации спанов")
    }
}