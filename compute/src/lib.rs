//! Библиотека вычислительного сервиса PII-модуля.
//!
//! Экспортирует модули, используемые как бинарём, так и интеграционными
//! тестами.

// Заглушки используют `todo!()`, который возвращает `!`; для `impl Trait`
// и коллекций это требует fallback на `()`. Разрешаем до реализации тел.
#![allow(dependency_on_unit_never_type_fallback)]

pub mod app_state;
pub mod config;
pub mod detect;
pub mod keys;
pub mod logging;
pub mod mask;
pub mod models;
pub mod pipeline;
pub mod ratelimit;
pub mod redis_client;
pub mod repo;
pub mod routers;
pub mod stats;
pub mod store;