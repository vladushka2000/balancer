//! Тесты семафора ограничения одновременных обработок.

use std::time::Duration;

use compute::ratelimit::ConcurrencySemaphore;

/// Проверяет, что семафор пропускает ровно `capacity` одновременных
/// захватов, а следующий ожидает освобождения слота.
#[tokio::test]
async fn test_semaphore_capacity() {
    let sem = ConcurrencySemaphore::new(2, Duration::from_millis(500));

    let first = sem.acquire().await;
    let second = sem.acquire().await;
    assert!(first.is_some(), "первый захват должен пройти");
    assert!(second.is_some(), "второй захват должен пройти");

    // Третий захват должен ждать, пока не освободится слот.
    let third = sem.acquire().await;
    assert!(third.is_none(), "третий захват должен упереться в лимит");
}

/// Проверяет, что при занятом слоте второй захват возвращает `None`
/// примерно через время ожидания.
#[tokio::test]
async fn test_semaphore_wait_timeout() {
    let sem = ConcurrencySemaphore::new(1, Duration::from_millis(100));

    let first = sem.acquire().await;
    assert!(first.is_some(), "первый захват должен пройти");

    let start = std::time::Instant::now();
    let second = sem.acquire().await;
    let elapsed = start.elapsed();

    assert!(second.is_none(), "второй захват должен вернуть None");
    assert!(
        elapsed >= Duration::from_millis(90),
        "ожидание должно занять ~100ms, прошло {elapsed:?}"
    );
}

/// Проверяет, что после освобождения слота захват снова проходит.
#[tokio::test]
async fn test_semaphore_release() {
    let sem = ConcurrencySemaphore::new(1, Duration::from_millis(100));

    let first = sem.acquire().await;
    assert!(first.is_some(), "первый захват должен пройти");

    // Слот освобождается автоматически при drop permit.
    drop(first);

    let second = sem.acquire().await;
    assert!(second.is_some(), "после освобождения захват должен пройти");
}