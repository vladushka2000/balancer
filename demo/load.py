#!/usr/bin/env python3
"""Генератор нагрузки PII-модуля под профиль жюри.

Профиль jury: разгон, среднее ~330 RPS, пики до 1000. 429 не ошибка,
но попадает в отчёт. На каждом payload_id: POST mask, затем POST demask
с НАШИМ result (жюри: «всегда ваша маска»).

Считает: RPS факт, latency mean/p50/p95/p99, долю 429, mask_ok, demask_ok
(должны совпасть), число roundtrip fail.

Опционально отдельный прогон --rps 2000 как плюс ТЗ, не gate.

Запуск: python demo/load.py --url http://localhost:8080 --profile jury
"""

import argparse
import json
import random
import statistics
import sys
import threading
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from typing import Any

PII_TEMPLATES = [
    "Клиент {fio}, паспорт {passport}",
    "Дата рождения {date}",
    "Место рождения: {city}",
    "Гражданство: Российская Федерация",
    "Выдан УФМС России, код подразделения {dept}",
    "Дата выдачи {date}",
    "Водительское удостоверение {dl}",
    "Адрес: {city}, ул. {street}, д. {house}",
    "Email {email}",
    "Телефон {phone}",
    "ИНН {inn}",
    "Карта {card}",
    "CVV {cvv}",
    "Держатель карты {fio}",
    "ПАСПОРТ {passport}",
    "Карта {card}, пин-код {pin}",
]

FAMILIES = ["Иванов", "Петров", "Сидоров", "Кузнецов", "Смирнов", "Волков"]
NAMES = ["Иван", "Пётр", "Алексей", "Дмитрий", "Сергей", "Николай"]
PATRONYMICS = ["Иванович", "Петрович", "Алексеевич", "Дмитриевич", "Сергеевич"]
CITIES = ["Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань"]
STREETS = ["Тверская", "Невский", "Ленина", "Пушкина", "Советская"]


def random_fio() -> str:
    return f"{random.choice(FAMILIES)} {random.choice(NAMES)} {random.choice(PATRONYMICS)}"


def random_passport() -> str:
    return f"{random.randint(1000, 9999)} {random.randint(100000, 999999)}"


def random_date() -> str:
    return f"{random.randint(1, 28):02d}.{random.randint(1, 12):02d}.{random.randint(1960, 2005)}"


def random_dept() -> str:
    return f"{random.randint(100, 999)}-{random.randint(100, 999)}"


def random_dl() -> str:
    return f"{random.randint(1, 99):02d} {random.choice(['АА', 'ВВ', 'СС', 'КК'])} {random.randint(100000, 999999)}"


def random_phone() -> str:
    return f"+7 9{random.randint(10, 99)} {random.randint(100, 999)}-{random.randint(10, 99)}-{random.randint(10, 99)}"


def random_email() -> str:
    return f"{random.choice(NAMES).lower()}.{random.choice(FAMILIES).lower()}@bank.ru"


def random_inn() -> str:
    return "7707083893"


def random_card() -> str:
    return "4111 1111 1111 1111"


def random_cvv() -> str:
    return str(random.randint(100, 999))


def random_pin() -> str:
    return f"{random.randint(1000, 9999)}"


def random_house() -> str:
    return str(random.randint(1, 200))


def generate_payload() -> str:
    template = random.choice(PII_TEMPLATES)
    return template.format(
        fio=random_fio(),
        passport=random_passport(),
        date=random_date(),
        city=random.choice(CITIES),
        dept=random_dept(),
        dl=random_dl(),
        street=random.choice(STREETS),
        house=random_house(),
        email=random_email(),
        phone=random_phone(),
        inn=random_inn(),
        card=random_card(),
        cvv=random_cvv(),
        pin=random_pin(),
    )


def post_process(url: str, payload: str, payload_id: str) -> tuple[int, str]:
    body = json.dumps({"payload": payload, "payload_id": payload_id}).encode("utf-8")
    request = urllib.request.Request(
        url + "/process",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            return response.status, json.loads(response.read().decode("utf-8")).get("result", "")
    except urllib.error.HTTPError as exc:
        return exc.code, ""


class LoadResult:
    def __init__(self) -> None:
        self.lock = threading.Lock()
        self.latencies: list[float] = []
        self.mask_ok = 0
        self.demask_ok = 0
        self.count_429 = 0
        self.roundtrip_fail = 0
        self.requests = 0


def run_pair(url: str, payload: str, payload_id: str, result: LoadResult) -> None:
    start = time.perf_counter()
    mask_status, mask = post_process(url, payload, payload_id)
    mask_latency = (time.perf_counter() - start) * 1000
    with result.lock:
        result.requests += 1
        result.latencies.append(mask_latency)
        if mask_status == 429:
            result.count_429 += 1
        elif mask_status == 200:
            result.mask_ok += 1

    start = time.perf_counter()
    demask_status, demask = post_process(url, mask, payload_id)
    demask_latency = (time.perf_counter() - start) * 1000
    with result.lock:
        result.requests += 1
        result.latencies.append(demask_latency)
        if demask_status == 429:
            result.count_429 += 1
        elif demask_status == 200:
            result.demask_ok += 1
            if demask != payload:
                result.roundtrip_fail += 1


class RateLimiter:
    def __init__(self, rate: float) -> None:
        self.rate = rate
        self.interval = 1.0 / rate if rate > 0 else 0.0
        self.next_time = time.perf_counter()
        self.lock = threading.Lock()

    def wait(self) -> None:
        if self.interval <= 0:
            return
        with self.lock:
            now = time.perf_counter()
            wait = self.next_time - now
            if wait > 0:
                time.sleep(wait)
            self.next_time = max(now, self.next_time) + self.interval


def ramp_profile(duration: float, avg_rps: float, peak_rps: float) -> list[tuple[float, float]]:
    """Возвращает список (время_от_старта, rps) для разгона 50 → avg → пики → спад."""
    points: list[tuple[float, float]] = []
    ramp_up = duration * 0.3
    peak_phase = duration * 0.4
    ramp_down = duration * 0.3
    steps = 60
    for i in range(steps):
        t = duration * i / steps
        if t < ramp_up:
            rps = 50 + (avg_rps - 50) * (t / ramp_up)
        elif t < ramp_up + peak_phase:
            phase = (t - ramp_up) / peak_phase
            rps = avg_rps + (peak_rps - avg_rps) * (0.5 + 0.5 * math_sin(phase * 4 * 3.14159))
        else:
            rps = avg_rps * (1 - (t - ramp_up - peak_phase) / ramp_down)
        points.append((t, max(10, rps)))
    return points


def math_sin(x: float) -> float:
    import math

    return math.sin(x)


def run_load(url: str, pairs: int, duration: float, avg_rps: float, peak_rps: float) -> LoadResult:
    result = LoadResult()
    profile = ramp_profile(duration, avg_rps, peak_rps)
    payloads = [(generate_payload(), f"load-{i}") for i in range(pairs)]

    start_time = time.perf_counter()
    profile_idx = 0
    with ThreadPoolExecutor(max_workers=64) as pool:
        futures = []
        while time.perf_counter() - start_time < duration:
            elapsed = time.perf_counter() - start_time
            while profile_idx < len(profile) - 1 and profile[profile_idx + 1][0] <= elapsed:
                profile_idx += 1
            target_rps = profile[profile_idx][1]
            limiter = RateLimiter(target_rps)
            batch = 8
            for _ in range(batch):
                if len(futures) >= 512:
                    done, futures = wait_any(futures)
                    for f in done:
                        f.result()
                payload, pid = payloads[random.randrange(len(payloads))]
                futures.append(pool.submit(run_pair, url, payload, pid, result))
                limiter.wait()
        for f in futures:
            f.result()
    return result


def wait_any(futures: list[Any]) -> tuple[list[Any], list[Any]]:
    import concurrent.futures

    done, pending = concurrent.futures.wait(futures, return_when=concurrent.futures.FIRST_COMPLETED)
    return list(done), list(pending)


def report(result: LoadResult, duration: float) -> None:
    lat = result.latencies
    mean = statistics.mean(lat) if lat else 0.0
    p50 = statistics.median(lat) if lat else 0.0
    p95 = percentile(lat, 95) if lat else 0.0
    p99 = percentile(lat, 99) if lat else 0.0
    rps = result.requests / duration if duration > 0 else 0.0
    print("=== LOAD REPORT ===")
    print(f"requests: {result.requests}")
    print(f"rps_actual: {rps:.1f}")
    print(f"latency_mean_ms: {mean:.2f}")
    print(f"latency_p50_ms: {p50:.2f}")
    print(f"latency_p95_ms: {p95:.2f}")
    print(f"latency_p99_ms: {p99:.2f}")
    print(f"count_429: {result.count_429}")
    print(f"mask_ok: {result.mask_ok}")
    print(f"demask_ok: {result.demask_ok}")
    print(f"roundtrip_fail: {result.roundtrip_fail}")
    ok = (
        result.mask_ok == result.demask_ok
        and result.roundtrip_fail == 0
        and p99 <= 1000
    )
    print("LOAD OK" if ok else "LOAD FAIL")
    return ok


def percentile(values: list[float], p: float) -> float:
    if not values:
        return 0.0
    sorted_vals = sorted(values)
    idx = min(len(sorted_vals) - 1, int(len(sorted_vals) * p / 100))
    return sorted_vals[idx]


def main() -> int:
    parser = argparse.ArgumentParser(description="Генератор нагрузки PII-модуля")
    parser.add_argument("--url", default="http://localhost:8080", help="Базовый URL сервиса")
    parser.add_argument("--profile", choices=["jury", "rps2000"], default="jury")
    parser.add_argument("--pairs", type=int, default=2000, help="Число пар payload_id")
    parser.add_argument("--duration", type=float, default=60.0, help="Длительность прогона, сек")
    args = parser.parse_args()
    url = args.url.rstrip("/")

    if args.profile == "jury":
        avg_rps, peak_rps = 330.0, 1000.0
    else:
        avg_rps, peak_rps = 1500.0, 2000.0

    result = run_load(url, args.pairs, args.duration, avg_rps, peak_rps)
    ok = report(result, args.duration)
    return 0 if ok else 1


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (urllib.error.URLError, urllib.error.HTTPError) as exc:
        print(f"connection error: {exc}")
        sys.exit(1)