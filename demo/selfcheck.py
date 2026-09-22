#!/usr/bin/env python3
"""Самопроверка PII-модуля: покрытие типов ТЗ, roundtrip 100%, FP.

Чекер закрыт — это единственная обратная связь до сдачи.

Gate (must):
1. Покрытие типов ТЗ: ФИО, ДР, место рождения, паспорт, гражданство,
   орган выдачи, код подразделения, дата выдачи, в/у, адрес (компоненты),
   email, телефон, ИНН, карта, CVV, ПИН, держатель карты.
   POST /process → в result ПД нет (или маска), types залогированы.
2. Roundtrip 100%: demask(наш result) == original. mask_ok == demask_ok.
3. FP: «Пушкин», адрес отделения банка — не маскируются.
4. Голая дата без маркера — не маскируется; «дата рождения …» — маскируется.
5. ПИН без карты — не маскируется (плюс ТЗ).
6. Регистр: «ПАСПОРТ 4509 123456» находится.

Регрессия стиля (не gate): похожесть на свои же прошлые маски.

Запуск: python demo/selfcheck.py --url http://localhost:8080
"""

import argparse
import json
import sys
import time
import urllib.error
import urllib.request
from typing import Any


def post_process(url: str, payload: str, payload_id: str) -> dict[str, Any]:
    """Отправляет запрос на /process и возвращает JSON-ответ."""
    body = json.dumps({"payload": payload, "payload_id": payload_id}).encode("utf-8")
    request = urllib.request.Request(
        url + "/process",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        return json.loads(response.read().decode("utf-8"))


def mask_and_demask(url: str, payload: str, payload_id: str) -> tuple[bool, str, str]:
    """Прогоняет пару mask→demask и возвращает (ok, mask, original)."""
    mask_resp = post_process(url, payload, payload_id)
    mask = mask_resp.get("result", "")
    demask_resp = post_process(url, mask, payload_id)
    demask = demask_resp.get("result", "")
    return demask == payload, mask, demask


def check_roundtrip(url: str, cases: list[tuple[str, str]]) -> tuple[int, int]:
    """Проверяет roundtrip для списка (payload, payload_id) и возвращает (ok, total)."""
    ok = 0
    for payload, payload_id in cases:
        success, _, _ = mask_and_demask(url, payload, payload_id)
        if success:
            ok += 1
    return ok, len(cases)


def check_no_pii(url: str, cases: list[tuple[str, str]]) -> tuple[int, int]:
    """Проверяет, что FP-кейсы не маскируются (mask == original), и возвращает (ok, total).

    Каждый кейс прогоняется как mask→demask, чтобы не нарушать pair invariant
    (mask_ok == demask_ok): маска без демаски раздувала бы mask_ok.
    """
    ok = 0
    for payload, payload_id in cases:
        success, mask, _ = mask_and_demask(url, payload, payload_id)
        if success and mask == payload:
            ok += 1
    return ok, len(cases)


def check_masked(url: str, cases: list[tuple[str, str]]) -> tuple[int, int]:
    """Проверяет, что ПД-кейсы маскируются (mask != original), и возвращает (ok, total).

    Каждый кейс прогоняется как mask→demask, чтобы не нарушать pair invariant.
    """
    ok = 0
    for payload, payload_id in cases:
        success, mask, _ = mask_and_demask(url, payload, payload_id)
        if success and mask != payload:
            ok += 1
    return ok, len(cases)


def get_stats(url: str) -> dict[str, Any]:
    """Возвращает JSON-ответ /stats."""
    request = urllib.request.Request(url + "/stats", method="GET")
    with urllib.request.urlopen(request, timeout=10) as response:
        return json.loads(response.read().decode("utf-8"))


def check_pair_invariant(url: str) -> tuple[bool, str]:
    """Проверяет mask_ok == demask_ok и что типы залогированы в /stats."""
    stats = get_stats(url)
    if stats.get("mask_ok") != stats.get("demask_ok"):
        return False, f"mask_ok={stats.get('mask_ok')} != demask_ok={stats.get('demask_ok')}"
    if not stats.get("detections_by_type"):
        return False, "detections_by_type пуст — типы не залогированы"
    return True, "pair invariant ok, types logged"


def main() -> int:
    """Запускает самопроверку и возвращает код выхода."""
    parser = argparse.ArgumentParser(description="Самопроверка PII-модуля")
    parser.add_argument("--url", default="http://localhost:8080", help="Базовый URL сервиса")
    args = parser.parse_args()
    url = args.url.rstrip("/")

    run_id = str(int(time.time() * 1000))

    pii_cases = [
        ("Клиент Иванов Иван Иванович, паспорт 4509 123456", f"pii-{run_id}-1"),
        ("Дата рождения 12.01.1990", f"pii-{run_id}-2"),
        ("Место рождения: Москва", f"pii-{run_id}-3"),
        ("Гражданство: Российская Федерация", f"pii-{run_id}-4"),
        ("Выдан УФМС России, код подразделения 123-456", f"pii-{run_id}-5"),
        ("Дата выдачи 15.03.2015", f"pii-{run_id}-6"),
        ("Водительское удостоверение 77 АА 123456", f"pii-{run_id}-7"),
        ("Адрес: Москва, ул. Тверская, д. 1", f"pii-{run_id}-8"),
        ("Email ivan.ivanov@bank.ru", f"pii-{run_id}-9"),
        ("Телефон +7 912 345-67-89", f"pii-{run_id}-10"),
        ("ИНН 7707083893", f"pii-{run_id}-11"),
        ("Карта 4111 1111 1111 1111", f"pii-{run_id}-12"),
        ("CVV 123", f"pii-{run_id}-13"),
        ("Держатель карты Иванов Иван Иванович", f"pii-{run_id}-15"),
        ("ПАСПОРТ 4509 123456", f"pii-{run_id}-16"),
        ("Карта 4111 1111 1111 1111, пин-код 1234", f"pii-{run_id}-17"),
    ]
    fp_cases = [
        ("Пушкин", f"fp-{run_id}-1"),
        ("Отделение банка по адресу: Москва, ул. Тверская, д. 1", f"fp-{run_id}-2"),
        ("Встреча 12.01.2024", f"fp-{run_id}-3"),
        ("ПИН-код 1234", f"fp-{run_id}-4"),
    ]

    masked_ok, masked_total = check_masked(url, pii_cases)
    roundtrip_ok, roundtrip_total = check_roundtrip(url, pii_cases)
    fp_ok, fp_total = check_no_pii(url, fp_cases)
    time.sleep(2.5)
    pair_ok, pair_msg = check_pair_invariant(url)

    print(f"masked: {masked_ok}/{masked_total}")
    print(f"roundtrip: {roundtrip_ok}/{roundtrip_total}")
    print(f"fp: {fp_ok}/{fp_total}")
    print(f"pair: {pair_msg}")

    if (
        masked_ok == masked_total
        and roundtrip_ok == roundtrip_total
        and fp_ok == fp_total
        and pair_ok
    ):
        print("SELFCHECK OK")
        return 0
    print("SELFCHECK FAIL")
    return 1


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (urllib.error.URLError, urllib.error.HTTPError) as exc:
        print(f"connection error: {exc}")
        sys.exit(1)