//! Модуль детекции ПД.
//!
//! Содержит подмодули структурной, NER и контекстной детекции, а также
//! слияние спанов.

pub mod base;
pub mod context;
pub mod merge;
pub mod ner;
pub mod structural;