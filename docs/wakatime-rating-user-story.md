# User Story: WakaTime Rating Plugin

| | |
|---|---|
| **Статус** | Draft |
| **Дата** | 2026-05-01 |
| **Стек** | gold-apisrv (Go, zenrpc v2, go-pg, PostgreSQL, Echo, appkit) |
| **Скоуп** | MVP — только рейтинг |

---

## 1. Контекст и мотивация

В команде нужен публичный лидерборд активности по данным WakaTime: топ участников за неделю, месяц и за всё время. Цели — прозрачность активности и лёгкая геймификация. Референс по содержанию — таблица с колонками `Rank / Username / Hours Coded / Daily Average`. **Дизайн UI вне скоупа этой истории** — речь только про backend-контракт и фоновую синхронизацию.

Плагин встраивается в существующий сервис `gold-apisrv` (vmkteam «gold» template) и публикует методы в публичный JSON-RPC namespace `/v1/rpc/`.

## 2. Пользователь

Ролей нет — единый сценарий для всех посетителей сайта. Любой пользователь может одновременно:
- видеть публичный рейтинг,
- закинуть свой WakaTime токен через форму на странице и попасть в этот же рейтинг.

Никаких авторизаций, личных кабинетов, флагов «я зарегистрирован» — только общий список и форма ввода.

### Форма добавления (UI-контракт)

Два обязательных поля:

| Поле | Подпись | Подсказка | Валидация |
|---|---|---|---|
| `username` | `Username (По желанию — ФИО)` | — | непустое, ≤ 100 символов |
| `secretApiKey` | `Secret API Key (Settings -> Account)` | `Начинается с waka_` | непустое, должно начинаться с `waka_` |

После сабмита форма дёргает `wakatime.register` (см. §5.1).

## 3. Истории

- **US-1.** Я хочу видеть топ всех участников за неделю / месяц / всё время, чтобы сравнить активность команды.
- **US-2.** Я хочу через форму на странице ввести свой username и WakaTime API key и попасть в общий рейтинг, чтобы моя статистика учитывалась наравне с другими.
- **US-3.** Я хочу, чтобы данные обновлялись сами раз в 5 минут, чтобы рейтинг был актуальным без ручных действий.

## 4. Функциональные требования

1. Поддерживаются три временны́х окна:
   - `week` — последние 7 календарных дней включая сегодня;
   - `month` — последние 30 календарных дней;
   - `all` — с момента регистрации участника в системе.
2. Сортировка топа — по суммарному времени программирования по убыванию.
3. Поля строки рейтинга:
   - `rank` (1-based, по убыванию `total_seconds`);
   - `username` (как ввёл пользователь в форме);
   - `total_seconds` + форматированная строка `"NN hrs MM mins"`;
   - `daily_average_seconds` + форматированная строка.
4. Фоновый синк агрегатов из WakaTime — раз в 5 минут для всех активных записей.
5. При недоступности WakaTime API возвращаем последний успешный кэш и проставляем флаг `stale_at` (timestamp последнего успешного синка для участника / для рейтинга в целом).
6. Если WakaTime отвечает `401/403` на токен, запись переводится в `disabled`:
   - синк по нему перестаёт ходить;
   - данные в рейтинге остаются (исторический след сохраняется).
7. `wakatime.register` валидирует токен реальным запросом к WakaTime (`/users/current`). Невалидный токен — отказ, запись не создаётся.
8. Уникальность по WakaTime-аккаунту: повторная отправка формы с тем же WakaTime-аккаунтом обновляет `username` и токен (если новый токен валиден). Дубликатов в рейтинге быть не должно.
9. Поле `secretApiKey` обязано начинаться с префикса `waka_` — иначе форма/RPC возвращают `validation_error` ещё до похода во внешний API.

## 5. Публичное JSON-RPC API

Namespace: `wakatime`. Адрес: `POST /v1/rpc/`. Авторизация не требуется.

### 5.1. `wakatime.register`

Принимает данные с формы. Оба поля обязательны.

**Request**
```json
{
  "jsonrpc": "2.0",
  "method": "wakatime.register",
  "params": {
    "username": "Иван Иванов",
    "secretApiKey": "waka_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  },
  "id": 1
}
```

**Response (success)**
```json
{
  "jsonrpc": "2.0",
  "result": { "id": 42, "username": "Иван Иванов", "status": "enabled" },
  "id": 1
}
```

**Ошибки**
- `1001 invalid_token` — WakaTime отверг ключ.
- `1002 wakatime_unavailable` — внешний API недоступен, попробовать позже.
- `1003 validation_error` — `username` пустой / `secretApiKey` пустой или не начинается с `waka_`.

### 5.2. `wakatime.top`

Топ участников.

**Request**
```json
{
  "jsonrpc": "2.0",
  "method": "wakatime.top",
  "params": { "period": "week", "limit": 50, "offset": 0 },
  "id": 2
}
```

`period` — `"week" | "month" | "all"`. `limit` по умолчанию 50, максимум 200. `offset` по умолчанию 0.

**Response**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "period": "week",
    "periodStart": "2026-04-25",
    "periodEnd": "2026-05-01",
    "totalSeconds": 449460,
    "totalHuman": "124 hrs 51 mins",
    "items": [
      {
        "rank": 1,
        "username": "Menemi",
        "totalSeconds": 104820,
        "totalHuman": "29 hrs 7 mins",
        "dailyAverageSeconds": 20940,
        "dailyAverageHuman": "5 hrs 49 mins",
        "staleAt": null
      }
    ],
    "fetchedAt": "2026-05-01T12:35:00Z"
  },
  "id": 2
}
```

## 6. Логическая модель данных

Без SQL-DDL — это пойдёт в MFD-генератор на этапе реализации.

### `wakatime_user`
| Поле | Тип | Назначение |
|---|---|---|
| `id` | int PK | |
| `username` | text | как ввёл пользователь в форме (ФИО или ник) |
| `wakatime_login` | text unique | логин из WakaTime (`/users/current`) — ключ уникальности записи |
| `wakatime_token` | bytea | зашифрованный API key |
| `status_id` | int | `1=enabled / 2=disabled` (как в шаблоне) |
| `created_at` | timestamptz | |
| `last_synced_at` | timestamptz null | время последнего успешного синка |
| `last_sync_error` | text null | последняя ошибка синка |

### `wakatime_stat`
| Поле | Тип | Назначение |
|---|---|---|
| `id` | bigint PK | |
| `wakatime_user_id` | int FK | |
| `period` | text | `week / month / all` |
| `period_start` | date | |
| `period_end` | date | |
| `total_seconds` | bigint | |
| `daily_average_seconds` | int | |
| `fetched_at` | timestamptz | |

Уникальный ключ: `(wakatime_user_id, period, period_end)` — по одной строке на участника на дату среза. История накапливается (для будущих графиков).

Топ строится `SELECT … DISTINCT ON (wakatime_user_id) …` по последнему `fetched_at` среди активных участников.

## 7. Фоновая синхронизация (cron, 5 минут)

- Использовать `vmkteam/cron` — добавить компонент в `pkg/app`.
- Job `wakatime.sync`:
  1. выбрать всех `enabled` участников;
  2. для каждого дёрнуть WakaTime: `/users/current/stats/last_7_days`, `/last_30_days`, `/all_time_since_today` (или эквивалент через `/summaries`);
  3. записать три новые строки в `wakatime_stat` (по одной на period);
  4. обновить `last_synced_at`, очистить `last_sync_error`;
  5. при `401/403` — перевести участника в `disabled`, сохранить причину.
- Лочить через `db.RunInLock(ctx, "wakatime.sync", …)` — на случай нескольких реплик.
- Per-user тайм-аут на запрос к WakaTime, чтобы один зависший участник не ронял всю джобу.
- Метрики Prometheus: `wakatime_sync_total{result="ok|error|disabled"}`, `wakatime_sync_duration_seconds`, `wakatime_sync_users_active`.

## 8. Acceptance criteria

- [ ] `wakatime.register` отказывает при невалидном токене и не пишет запись в БД.
- [ ] `wakatime.register` отказывает на `secretApiKey` без префикса `waka_` без обращения к WakaTime.
- [ ] Повторная отправка формы с тем же WakaTime-аккаунтом не создаёт дубликат, а обновляет `username` и токен существующей записи.
- [ ] `wakatime.top` корректно сортирует по `total_seconds` и возвращает корректные `rank`.
- [ ] Значения времени совпадают с цифрами в личном кабинете WakaTime ±5 минут.
- [ ] При перезапуске сервиса топ доступен сразу (данные читаются из БД, не из памяти).
- [ ] Падение WakaTime API не валит сервис: `wakatime.top` отдаёт последний кэш с заполненным `staleAt`.
- [ ] Лаг между WakaTime и локальными данными ≤ 5 минут при штатной работе.
- [ ] Токен хранится зашифрованным; в логах и в ответах API он не появляется.
- [ ] Методы описаны `//zenrpc:` доккомментариями и видны в `/v1/rpc/doc/` (SMD) и `openrpc.json`.
- [ ] Cron-джоба зарегистрирована в `MetadataOpts` (`/debug/metadata`).

## 9. Безопасность и приватность

- WakaTime token — секрет уровня пароля. Шифровать AES-GCM ключом из конфига (`Wakatime.EncryptionKey`).
- Никогда не логировать токен и тело ответа WakaTime «как есть».
- Rate-limit на `wakatime.register` (например, 5 попыток на IP в минуту) — защита от перебора чужих токенов.
- Не возвращать токен наружу ни одним методом.

## 10. Out of scope (следующие итерации)

- Дизайн страницы и оформление формы — описан только UI-контракт (имена и валидация полей).
- Coding Now, фильтр по проекту, текущий проект участника.
- Графики истории и «Топ недель».
- Админский CRUD участников через `/v1/vt/` (отвязка, смена токена, бан).
- Командные/групповые рейтинги.
- Личные кабинеты и любые виды авторизации.

## 11. Технические заметки для разработчика

- Сервис: `pkg/rpc/wakatime_service.go` + `pkg/rpc/wakatime_model.go` (по аналогии с `pkg/vt/vt_service.go` / `pkg/vt/vt_model.go`).
- Регистрация в `pkg/rpc/server.go` → `RegisterAll`.
- WakaTime HTTP-клиент: новый пакет `pkg/wakatime/` поверх `appkit` HTTP-клиента (метрики, X-Request-ID).
- Конфиг: новая секция `[Wakatime]` в `cfg/local.toml.dist` — `BaseURL`, `Timeout`, `EncryptionKey`, `SyncInterval` (default `5m`).
- Схема БД: добавить таблицы в `docs/wakatime.sql`, синхронизировать MFD: `make mfd-xml` → `make mfd-model` → `make mfd-repo NS=common`.
- Cron — через `vmkteam/cron`, монтируется в `pkg/app/app.go` рядом с остальными компонентами.
- Метаданные — дополнить `MetadataOpts` в `pkg/app/app.go` (новый async-сервис `wakatime.sync`).
