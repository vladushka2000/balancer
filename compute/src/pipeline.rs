//! Конвейер обработки: детекция → контекст → слияние → маскирование.

use crate::detect::base::DetectorRegistry;
use crate::detect::context::ContextRule;
use crate::detect::ner::NERDetector;
use crate::mask::masker::Masker;

/// Конвейер обработки текста.
pub struct DetectionPipeline {
    /// Реестр структурных детекторов.
    structural: DetectorRegistry,
    /// NER-детектор.
    ner: NERDetector,
    /// Контекстное правило.
    context: ContextRule,
    /// Маскер.
    masker: Masker,
}

impl DetectionPipeline {
    /// Обрабатывает текст: возвращает маскированный текст и типы ПД.
    pub async fn process(&self, text: &str) -> (String, Vec<String>) {
        todo!("реализация конвейера обработки")
    }
}