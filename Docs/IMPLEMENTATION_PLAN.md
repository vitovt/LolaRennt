# Implementation Plan

## 1. Мета

Цей документ описує послідовний план реалізації `SPEC v1.0` для desktop-застосунку на `Go + Fyne`.

План побудований так, щоб:

* якомога раніше отримати робочий end-to-end pipeline;
* мінімізувати ризик розходження між preview та export;
* винести складні залежності в окремі етапи;
* залишити можливість стабілізації після кожної фази.

---

## 2. Принципи реалізації

Ключові технічні принципи:

* один спільний renderer core для preview та export;
* project state не повинен залежати від UI;
* анімація має бути детермінованою при однаковому `seed`;
* export pipeline має працювати через покадровий render у `PNG`;
* інтеграція з `ffmpeg` у `v1.0` обмежується перевіркою наявності і генерацією готової команди;
* кожна фаза повинна завершуватися робочим, перевірним станом.

---

## 3. Цільова структура модулів

Рекомендовані пакети:

* `project`
* `model`
* `charset`
* `glyphs`
* `layout`
* `animation`
* `renderer`
* `preview`
* `background`
* `export`
* `ffmpeg`
* `storage`
* `presets`
* `ui`

Мінімальний розподіл відповідальності:

* `model`: структури даних і enums;
* `project`: orchestration поточного проєкту;
* `charset`: доступні символи, перевірка тексту, replacement rules;
* `glyphs`: сегментні та матричні glyph maps;
* `layout`: fixed-width layout, line breaking, alignment;
* `animation`: scramble state machine;
* `renderer`: покадровий рендер тексту і ефектів;
* `background`: color, gradient, image, video background;
* `preview`: live preview loop;
* `export`: render sequence, progress, cancel;
* `ffmpeg`: detect path, build command;
* `storage`: save/load project JSON;
* `presets`: style / animation / export presets;
* `ui`: вкладки, форми, binding, interaction.

---

## 4. Порядок реалізації

## 4.1. Фаза 0. Bootstrap

Мета:

* підготувати каркас проєкту;
* закласти структуру модулів;
* домовитися про модель даних.

Deliverables:

* початковий застосунок на `Fyne`;
* базова структура директорій і пакетів;
* `Project` model;
* `JSON` save/load каркас;
* базовий app state і event flow.

Критерій завершення:

* застосунок запускається;
* можна створити порожній проєкт;
* можна зберегти і відкрити тестовий `project.json`.

---

## 4.2. Фаза 1. Text + Charset foundation

Мета:

* закласти текстову модель і правила підтримки символів.

Deliverables:

* multi-line text model;
* language profiles:
  * English
  * German
  * Ukrainian
  * Russian
* комбінований charset builder;
* перевірка supported / unsupported символів;
* uppercase transform;
* базова пунктуація;
* UI для text input та charset settings.

Критерій завершення:

* користувач може ввести multi-line текст;
* unsupported символи виявляються коректно;
* стан зберігається в `project.json`.

---

## 4.3. Фаза 2. Glyphs + Static renderer

Мета:

* отримати перший візуальний результат без анімації.

Deliverables:

* glyph representation для:
  * `Block matrix`
  * `Dot-matrix`
  * `Segment display`
* automatic glyph preset selection для `v1.0`;
* static text renderer;
* inactive blocks / dots for matrix modes;
* main color;
* glow;
* fixed-width layout;
* line spacing, character spacing, alignment, padding;
* static preview у вкладці `Текст і стиль`.

Критерій завершення:

* preview коректно показує multi-line текст у трьох display modes;
* один і той самий input дає стабільний статичний результат.

---

## 4.4. Фаза 3. Animation core

Мета:

* реалізувати scramble pipeline, не прив'язаний до UI.

Deliverables:

* `seed` handling:
  * пусте поле -> авто-генерація;
  * manual override;
  * button `Generate random seed`;
* animation settings model;
* `Scramble basic`;
* `Scramble with lock`;
* order modes:
  * left-to-right
  * right-to-left
  * center-out
  * random
  * by lines
* lock modes:
  * hard lock
  * probabilistic lock
* frame-by-frame animation state calculation.

Критерій завершення:

* при однаковому `seed` анімація відтворюється однаково;
* анімаційний state можна обчислити для довільного кадру без UI.

---

## 4.5. Фаза 4. Live preview + playback

Мета:

* підключити animation core до UI і отримати робочий preview.

Deliverables:

* animated preview pane;
* play / pause / stop / restart / loop;
* timeline scrubber;
* current frame indicator;
* current time indicator;
* zoom;
* checkerboard preview mode;
* safe area toggle.

Критерій завершення:

* користувач може запустити і зупинити анімацію;
* preview використовує той самий renderer core, що і export pipeline.

---

## 4.6. Фаза 5. Background subsystem

Мета:

* додати всі фонові режими, потрібні для `v1.0`.

Deliverables:

* transparent background;
* solid color background;
* gradient background;
* image background:
  * file picker;
  * fit / fill / center / stretch;
  * opacity;
* video background:
  * file picker;
  * basic playback sync;
  * fit / fill / center / stretch;
* compositing тексту поверх background.

Критерій завершення:

* preview і export підтримують однакові режими фону;
* image і video background використовуються без окремого рендер-пайплайна для тексту.

Примітка:

* background video є найризикованішою частиною `v1.0`;
* якщо реалізація ускладнюється через декодування і синхронізацію, цю фазу варто робити після стабільного PNG pipeline.

---

## 4.7. Фаза 6. Export pipeline

Мета:

* довести продукт до першого production-like output.

Deliverables:

* export settings UI;
* size presets:
  * `1920x1080`
  * `1080x1920`
* custom size;
* fps;
* start / end frame;
* output folder;
* file prefix;
* overwrite policy;
* supersampling;
* render loop у `PNG sequence`;
* progress reporting;
* ETA;
* cancel;
* log output.

Критерій завершення:

* користувач може стабільно відрендерити повну послідовність кадрів;
* кадри збігаються з preview за логікою animation/render state.

---

## 4.8. Фаза 7. FFmpeg command builder

Мета:

* завершити video workflow `v1.0` без обов'язкового внутрішнього encode.

Deliverables:

* auto-detect `ffmpeg`;
* manual path override;
* generator готової `ffmpeg` команди;
* copy command;
* різні шаблони команди для:
  * transparent-oriented workflow через PNG sequence;
  * baked-background workflow для фінального відео.

Критерій завершення:

* користувач може відразу після export скопіювати валідну команду `ffmpeg` і зібрати відео поза застосунком.

---

## 4.9. Фаза 8. Projects + Presets

Мета:

* зробити інструмент придатним до повторного використання.

Deliverables:

* save / load project;
* recent projects;
* style presets;
* animation presets;
* export presets;
* metadata:
  * project name
  * notes
  * tags
  * created at
  * updated at

Критерій завершення:

* повний робочий стан проєкту відновлюється після повторного відкриття.

---

## 4.10. Фаза 9. Stabilization

Мета:

* прибрати технічний борг, не розширюючи scope.

Deliverables:

* bug fixing;
* performance tuning;
* memory usage review;
* error handling cleanup;
* compatibility pass для Linux / Windows;
* smoke tests на основні user flows.

Критерій завершення:

* `v1.0` проходить acceptance criteria зі `SPEC`.

---

## 5. Рекомендований порядок розробки по пріоритету

Реальний пріоритет виконання:

1. Фаза 0
2. Фаза 1
3. Фаза 2
4. Фаза 3
5. Фаза 4
6. Фаза 6
7. Фаза 7
8. Фаза 8
9. Фаза 5
10. Фаза 9

Причина такого порядку:

* спершу треба довести текст, glyphs, animation і export до стабільного стану;
* `background video` краще інтегрувати після того, як базовий renderer і PNG pipeline уже працюють;
* це зменшує ризик, що одна складна підсистема заблокує весь MVP.

---

## 6. Технічні рішення, які треба прийняти рано

До кінця Фази 1 бажано зафіксувати:

* формат `project.json`;
* enums для display mode, animation type, lock order, lock mode;
* підхід до glyph storage:
  * hardcoded Go maps
  * JSON assets
* базову модель frame timing;
* підхід до image loading;
* підхід до video decoding для background video.

До кінця Фази 2 бажано зафіксувати:

* coordinate system renderer;
* одиниці вимірювання layout;
* спосіб compositing glow;
* чи буде supersampling на рівні export renderer, чи через масштабування canvas.

---

## 7. Основні ризики

### 7.1. Background video

Ризики:

* декодування відео;
* синхронізація кадрів;
* продуктивність preview;
* залежності на зовнішні бібліотеки або `ffmpeg`.

Пом'якшення:

* спочатку завершити image background;
* побудувати API background source так, щоб image і video мали спільний інтерфейс;
* для `v1.0` не робити audio, trimming, color grading.

### 7.2. Glyph coverage

Ризики:

* неповне покриття українських і російських літер;
* складність сумісності segment-гліфів з усіма alphabet mixes.

Пом'якшення:

* зафіксувати мінімальний гарантований набір символів;
* явно показувати unsupported characters;
* для `v1.0` дозволити автоматичний fallback strategy.

### 7.3. Preview vs export drift

Ризик:

* preview і export почнуть використовувати різні code paths.

Пом'якшення:

* спільний animation core;
* спільний renderer core;
* export як frame iteration поверх тих самих state calculators.

---

## 8. Мінімальна стратегія тестування

### 8.1. Unit tests

Покрити:

* charset validation;
* glyph lookup;
* fixed-width layout calculations;
* deterministic animation by seed;
* frame index -> animation state;
* `ffmpeg` command generation;
* save / load project JSON.

### 8.2. Snapshot / golden tests

Бажано покрити:

* static render одного кадру;
* один і той самий `seed` дає той самий результат;
* transparent PNG render;
* baked background render.

### 8.3. Manual smoke tests

Перевірити:

* multi-line text;
* змішаний language set;
* `16:9` export;
* `9:16` export;
* image background;
* video background;
* repeatable export after reopen project.

---

## 9. Definition of Done для v1.0

Реалізацію можна вважати завершеною, якщо:

* усі acceptance criteria зі `SPEC v1.0` виконані;
* є робочий save/load project flow;
* є стабільний `PNG sequence` export;
* є валідна `ffmpeg` команда для video assembly;
* preview та export дають однаковий результат при однаковому `seed`;
* основні user flows перевірені вручну на Linux і Windows.
