//! Per-type правила маскирования ПД.

use std::collections::HashMap;

use crate::models::Span;

/// Маскирует паспорт: "4509 123456" → "45** ****56".
pub fn mask_passport(text: &str, span: &Span) -> String {
    todo!("реализация маскирования паспорта")
}

/// Маскирует ФИО: "Иванов Иван Иванович" → "И. И. И.".
pub fn mask_fio(text: &str, span: &Span) -> String {
    todo!("реализация маскирования ФИО")
}

/// Маскирует телефон: "+7 912 345-67-89" → "+7 9** ***-**-89".
pub fn mask_phone(text: &str, span: &Span) -> String {
    todo!("реализация маскирования телефона")
}

/// Маскирует email: "ivan@bank.ru" → "i*******@bank.ru".
pub fn mask_email(text: &str, span: &Span) -> String {
    todo!("реализация маскирования email")
}

/// Маскирует карту: "4111...1111" → "4111 **** **** 1111".
pub fn mask_card(text: &str, span: &Span) -> String {
    todo!("реализация маскирования карты")
}

/// Маскирует ИНН: "7707083893" → "77** ****** 93".
pub fn mask_inn(text: &str, span: &Span) -> String {
    todo!("реализация маскирования ИНН")
}

/// Маскирует СНИЛС: "123-456-789 01" → "123-***-*** **".
pub fn mask_snils(text: &str, span: &Span) -> String {
    todo!("реализация маскирования СНИЛС")
}

/// Маскирует дату: "12.01.1990" → "**.**.1990".
pub fn mask_date(text: &str, span: &Span) -> String {
    todo!("реализация маскирования даты")
}

/// Маскирует CVV: "123" → "***".
pub fn mask_cvv(text: &str, span: &Span) -> String {
    todo!("реализация маскирования CVV")
}

/// Маскирует PIN: "1234" → "***".
pub fn mask_pin(text: &str, span: &Span) -> String {
    todo!("реализация маскирования PIN")
}

/// Маскирует почтовый индекс: "123456" → "******".
pub fn mask_postal_code(text: &str, span: &Span) -> String {
    todo!("реализация маскирования почтового индекса")
}

/// Маскирует код подразделения: "123-456" → "***-***".
pub fn mask_department_code(text: &str, span: &Span) -> String {
    todo!("реализация маскирования кода подразделения")
}

/// Маскирует адрес: "Москва, ул. Тверская, д. 1" → "Москва, ул. ******, д. **".
pub fn mask_address(text: &str, span: &Span) -> String {
    todo!("реализация маскирования адреса")
}

/// Маскирует водительское удостоверение: "77 АА 123456" → "77** ****56".
pub fn mask_driver_license(text: &str, span: &Span) -> String {
    todo!("реализация маскирования водительского удостоверения")
}

/// Fallback-маскирование (partial).
pub fn mask_default(text: &str, span: &Span) -> String {
    todo!("реализация fallback-маскирования")
}

/// Возвращает карту «тип ПД → функция маскирования».
pub fn mask_rules() -> HashMap<&'static str, fn(&str, &Span) -> String> {
    todo!("реализация карты правил маскирования")
}