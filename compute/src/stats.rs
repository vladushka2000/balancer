//! Сбор и отдача статистики (RPS, latency, TPS).

use std::collections::{HashMap, VecDeque};
use std::sync::{Arc, Mutex};

/// Внутреннее состояние статистики.
struct StatsInner {
    /// Общее число обработанных запросов.
    requests_total: u64,
    /// Число детекций по типам ПД.
    detections_by_type: HashMap<String, u64>,
    /// Скользящее окно задержек (1000 последних).
    latencies: VecDeque<f64>,
}

/// Снимок статистики.
pub struct StatsSnapshot;

/// Сборщик статистики.
#[derive(Clone)]
pub struct Stats {
    /// Внутреннее состояние.
    inner: Arc<Mutex<StatsInner>>,
}

impl Stats {
    /// Записывает факт обработки запроса.
    pub fn record(&self, types: &[String], latency_ms: f64) {
        todo!("реализация записи статистики")
    }

    /// Возвращает снимок текущей статистики.
    pub fn snapshot(&self) -> StatsSnapshot {
        todo!("реализация снимка статистики")
    }
}
