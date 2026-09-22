//! Семафор ограничения одновременных обработок.
//!
//! Ограничивает число одновременно выполняемых запросов. Если слот не
//! освободился за заданное время ожидания, [`acquire`] возвращает `None`,
//! что позволяет сервису вернуть клиенту ответ 429/503 вместо бесконечного
//! ожидания.
//!
//! Возвращаемый permit удерживает слот до тех пор, пока жив; освобождение
//! происходит автоматически при его drop.

use std::sync::Arc;
use std::time::Duration;

use tokio::sync::{OwnedSemaphorePermit, Semaphore};

/// Семафор с ограничением по времени ожидания слота.
#[derive(Clone)]
pub struct ConcurrencySemaphore {
    /// Семафор с фиксированной ёмкостью.
    sem: Arc<Semaphore>,
    /// Максимальное время ожидания слота.
    wait: Duration,
}

impl ConcurrencySemaphore {
    /// Создаёт семафор с заданной ёмкостью и временем ожидания.
    ///
    /// # Аргументы
    ///
    /// * `capacity` — максимальное число одновременных слотов.
    /// * `wait` — максимальное время ожидания слота.
    pub fn new(capacity: usize, wait: Duration) -> Self {
        ConcurrencySemaphore {
            sem: Arc::new(Semaphore::new(capacity)),
            wait,
        }
    }

    /// Пытается захватить слот семафора.
    ///
    /// Возвращает `Some(permit)`, если слот получен в течение времени
    /// ожидания, и `None`, если время ожидания истекло. Полученный permit
    /// удерживает слот до своего drop.
    pub async fn acquire(&self) -> Option<OwnedSemaphorePermit> {
        match tokio::time::timeout(self.wait, self.sem.clone().acquire_owned()).await {
            Ok(Ok(permit)) => Some(permit),
            _ => None,
        }
    }
}
