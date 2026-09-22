//! Структурная детекция ПД: regex-детекторы с валидаторами checksum.

use crate::detect::base::{Detector, DetectorRegistry};
use crate::models::Span;

/// Детектор паспорта (type="passport").
pub struct PassportDetector;

impl Detector for PassportDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора паспорта")
    }
}

/// Детектор водительского удостоверения (type="driver_license").
pub struct DriverLicenseDetector;

impl Detector for DriverLicenseDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора водительского удостоверения")
    }
}

/// Детектор ИНН (type="inn").
pub struct INNDetector;

impl Detector for INNDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора ИНН")
    }
}

/// Детектор СНИЛС (type="snils").
pub struct SNILSDetector;

impl Detector for SNILSDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора СНИЛС")
    }
}

/// Детектор телефона (type="phone").
pub struct PhoneDetector;

impl Detector for PhoneDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора телефона")
    }
}

/// Детектор email (type="email").
pub struct EmailDetector;

impl Detector for EmailDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора email")
    }
}

/// Детектор банковской карты (type="card").
pub struct CardDetector;

impl Detector for CardDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора карты")
    }
}

/// Детектор CVV (type="cvv").
pub struct CVVDetector;

impl Detector for CVVDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора CVV")
    }
}

/// Детектор PIN-кода (type="pin").
pub struct PINDetector;

impl Detector for PINDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора PIN")
    }
}

/// Детектор даты (type="date").
pub struct DateDetector;

impl Detector for DateDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора даты")
    }
}

/// Детектор почтового индекса (type="postal_code").
pub struct PostalCodeDetector;

impl Detector for PostalCodeDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора почтового индекса")
    }
}

/// Детектор кода подразделения (type="department_code").
pub struct DepartmentCodeDetector;

impl Detector for DepartmentCodeDetector {
    fn detect(&self, text: &str) -> Vec<Span> {
        todo!("реализация детектора кода подразделения")
    }
}

/// Создаёт реестр со всеми структурными детекторами.
pub fn create_structural_registry() -> DetectorRegistry {
    todo!("реализация создания реестра структурных детекторов")
}