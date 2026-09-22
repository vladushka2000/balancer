//! Общее состояние приложения, передаваемое в обработчики.

use crate::pipeline::DetectionPipeline;
use crate::ratelimit::ConcurrencySemaphore;
use crate::repo::Repo;
use crate::stats::Stats;
use crate::store::CorrespondenceStore;

/// Общее состояние приложения.
pub struct AppState {
    /// Хранилище соответствий «оригинал ↔ маска».
    pub store: CorrespondenceStore,
    /// Конвейер обработки.
    pub pipeline: DetectionPipeline,
    /// Семафор ограничения одновременных обработок.
    pub semaphore: ConcurrencySemaphore,
    /// Сборщик статистики.
    pub stats: Stats,
    /// Репозиторий конфигураций систем.
    pub repo: Repo,
}