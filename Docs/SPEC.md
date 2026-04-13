# SPEC v1.0

## 1. Документ

Цей документ фіксує технічне завдання для першої робочої версії десктопного застосунку для створення текстової анімації у стилі segment / dot-matrix display.

Статус документа: `v1.0`

Мета документа:

* зафіксувати обсяг `v1.0`;
* прибрати відкриті питання з чернетки;
* дати розробнику один узгоджений орієнтир по функціональності, UI, збереженню даних і межах MVP.

---

## 2. Призначення продукту

Застосунок має дозволяти користувачу:

* ввести текст у кілька рядків;
* обрати тип візуалізації: `segment` або `dot-matrix`;
* налаштувати стиль, кольори, фон і параметри анімації;
* бачити preview;
* відтворювати scramble-анімацію;
* рендерити кадри у `PNG sequence`;
* отримувати готову команду `ffmpeg` для збирання відео з відрендерених кадрів.

Продукт орієнтований на титри, стилізовані текстові вставки, motion graphics overlays та монтажні заготовки у стилі електронних сегментних індикаторів.

---

## 3. Візуальний референс

Скріншоти:

* [screenshot1.jpeg](/mnt/vitovt/Desktop/Dev/LolaRennt/Docs/screenshot1.jpeg)
* [screenshot2.jpeg](/mnt/vitovt/Desktop/Dev/LolaRennt/Docs/screenshot2.jpeg)
* [screenshot3.jpeg](/mnt/vitovt/Desktop/Dev/LolaRennt/Docs/screenshot3.jpeg)

Вони задають художній напрямок, а не вимогу до pixel-perfect копіювання.

Для `v1.0` важливо відтворити такі риси:

* яскравий сегментний або матричний текст;
* переважно uppercase-подання;
* fixed-width характер набору;
* придатність для одного рядка і для multi-line композицій;
* робота як на чорному фоні, так і поверх зображення або відео;
* відчуття LED / VFD / digital display, а не звичайного шрифтового рендеру.

---

## 4. Технічні обмеження і стек

Для `v1.0` застосунок має бути реалізований як desktop app на:

* `Go`
* `Fyne`

Вимога: UI має бути суто на Go без web runtime.

Рекомендована внутрішня структура:

* окремий модуль стану проєкту;
* окремий модуль гліфів;
* окремий модуль layout;
* окремий модуль animation;
* окремий модуль renderer;
* окремий модуль export;
* окремий модуль генерації `ffmpeg` команди.

---

## 5. Платформи

Обов'язкові платформи `v1.0`:

* Linux
* Windows

`macOS` допускається як необов'язкова ціль після стабілізації `v1.0`.

---

## 6. Scope версій

### 6.1. MVP / v1.0

У `v1.0` обов'язково входить:

* desktop UI на `Go + Fyne`;
* multi-line текст;
* live preview;
* `segment` display;
* `dot-matrix` display;
* підтримка алфавітів:
  * English
  * German
  * Ukrainian
  * Russian
* базові знаки пунктуації;
* fixed-width layout;
* scramble-анімація;
* deterministic animation при зафіксованому `seed`;
* фон:
  * transparent
  * solid color
  * gradient
  * image
  * video
* export у `PNG sequence`;
* генерація готової `ffmpeg` команди для збирання відео;
* формати композиції:
  * `16:9`
  * `9:16`
* save / load project;
* presets для стилю, анімації та експорту.

### 6.2. v1.1

Можна винести в `v1.1`:

* one-click внутрішній encode у фінальне відео через `ffmpeg`;
* розширені glyph presets;
* додаткові animation presets;
* тонше керування background video;
* покращені export presets;
* autosave recovery.

### 6.3. Later

Пізніше:

* custom glyph editor;
* timeline keyframes;
* batch rendering;
* CLI render mode;
* browser version;
* MOV ProRes 4444;
* custom preset editor для random pools.

---

## 7. Основні сценарії користувача

### 7.1. Базовий сценарій

1. Користувач створює новий проєкт.
2. Вводить текст у кілька рядків.
3. Обирає профіль символів і тип індикатора.
4. Налаштовує стиль, кольори і layout.
5. Додає фон: прозорий, колір, градієнт, зображення або відео.
6. Налаштовує scramble-анімацію.
7. Перевіряє preview.
8. Рендерить `PNG sequence`.
9. Копіює або запускає згенеровану команду `ffmpeg` для збирання відео.
10. Зберігає проєкт.

### 7.2. Повторне відкриття

1. Користувач відкриває раніше збережений проєкт.
2. Бачить той самий текст, стиль, seed, фон і параметри експорту.
3. Повторний preview та export дають той самий результат.

---

## 8. Основні сутності

### 8.1. Project

`Project` містить:

* `TextSettings`
* `CharsetSettings`
* `DisplaySettings`
* `StyleSettings`
* `LayoutSettings`
* `AnimationSettings`
* `BackgroundSettings`
* `ExportSettings`
* `Metadata`

### 8.2. Charset Profile

Профіль допустимих символів:

* `English`
* `German`
* `Ukrainian`
* `Russian`
* комбіновані профілі:
  * `English + German`
  * `English + Ukrainian`
  * `English + Russian`
  * `German + Ukrainian`
  * `German + Russian`
  * `English + German + Ukrainian + Russian`

У `v1.0` достатньо реалізувати комбінований режим через один список доступних символів, який обчислюється з активних мов.

### 8.3. Display Mode

* `Segment`
* `Dot-matrix`

### 8.4. Glyph Strategy

У `v1.0` допустима найпростіша реалізація:

* користувач обирає `Display Mode`;
* система автоматично використовує найпростіший сумісний glyph preset для поточного набору мов і символів;
* ручний advanced override preset не є обов'язковим у `v1.0`.

### 8.5. Seed

`Seed` визначає псевдовипадкову послідовність анімації.

Правила для `v1.0`:

* якщо поле `Seed` порожнє, система автоматично генерує значення;
* автоматично згенерований `seed` зберігається у проєкті;
* є кнопка `Generate random seed`;
* після генерації або ручного введення значення `seed` більше не змінюється саме по собі;
* однаковий проєкт + однаковий `seed` мають давати однаковий результат у preview та export.

---

## 9. Вимоги до тексту і символів

Система повинна:

* підтримувати multi-line input;
* підтримувати uppercase workflow;
* перевіряти кожен символ на підтримку поточним профілем;
* підсвічувати unsupported characters;
* за потреби пропонувати replacement;
* підтримувати базову пунктуацію щонайменше:
  * `.`
  * `,`
  * `:`
  * `;`
  * `!`
  * `?`
  * `-`
  * `(`
  * `)`
  * `/`
* коректно працювати з німецькими умлаутами;
* коректно працювати з українськими та російськими літерами в межах заявленого набору гліфів.

---

## 10. Вимоги до layout

Для `v1.0` layout має бути `fixed-width`.

Система повинна підтримувати:

* character spacing;
* line spacing;
* alignment:
  * left
  * center
  * right
* padding;
* safe area preview;
* роботу з одним рядком і з кількома рядками.

Variable-width layout не входить у `v1.0`.

---

## 11. Вимоги до рендеру

Система повинна:

* рендерити символи не як системний шрифт, а як набір сегментів або матричних точок;
* використовувати внутрішнє представлення гліфа у вигляді bitmap / bitmask / аналогічної структури;
* підтримувати inactive segments / inactive dots;
* підтримувати glow;
* коректно працювати на прозорому фоні;
* давати однаковий результат у preview та export, окрім допустимих відмінностей anti-aliasing.

---

## 12. Вимоги до анімації

### 12.1. Обов'язкові можливості `v1.0`

* scramble-анімація;
* покадрове обчислення стану для кожного символу;
* перехід від випадкових символів до target text;
* lock logic;
* deterministic result при однаковому `seed`;
* однакові параметри анімації для preview та export.

### 12.2. Мінімальний набір режимів `v1.0`

Обов'язково:

* `Scramble basic`
* `Scramble with lock`

Опційно, якщо не ускладнює реалізацію:

* `Reveal then scramble then lock`

### 12.3. Мінімальні параметри `v1.0`

* total duration;
* intro delay;
* outro hold;
* per-character delay;
* random switch rate;
* seed;
* lock order;
* lock mode;
* allow empty cell;
* allow invalid random character.

### 12.4. Порядок і логіка lock

Потрібно підтримати:

* order:
  * left-to-right
  * right-to-left
  * center-out
  * random
  * by lines
* lock mode:
  * hard lock
  * probabilistic lock

### 12.5. Visual effects

У `v1.0` достатньо:

* glow intensity;
* inactive segment visibility;
* flicker amount;
* noise overlay;
* scanline effect як optional toggle.

---

## 13. Вимоги до background

### 13.1. Підтримувані режими `v1.0`

* transparent
* solid color
* gradient
* image
* video

### 13.2. Background image

Потрібно підтримати:

* file picker;
* fit / fill / center / stretch;
* opacity.

### 13.3. Background video

Для `v1.0` достатньо:

* вибір одного відеофайлу як background source;
* відтворення background video у preview;
* використання background video під час export;
* синхронізація background video з поточним таймінгом композиції;
* базові режими fit / fill / center / stretch.

Необов'язково для `v1.0`:

* trimming;
* time remap;
* loop editor;
* advanced color correction;
* audio.

---

## 14. Вимоги до export

### 14.1. Основний формат `v1.0`

Основний експорт `v1.0`:

* `PNG sequence`

Система повинна:

* експортувати нумеровані `PNG` кадри;
* підтримувати transparent export;
* підтримувати export з baked background;
* зберігати export settings у project file.

### 14.2. Відео для `v1.0`

У `v1.0` вбудований one-click encode не є обов'язковим.

Обов'язково:

* сформувати готову команду `ffmpeg` для складання відео з відрендереної послідовності;
* дати користувачу можливість скопіювати цю команду;
* врахувати у сформованій команді:
  * fps;
  * шаблон імені кадрів;
  * формат вихідного файлу;
  * фон уже baked у кадрах, якщо вибрано непрозорий режим.

### 14.3. Формати композиції

У `v1.0` потрібні щонайменше такі preset-формати:

* `1920x1080` (`16:9`)
* `1080x1920` (`9:16`)

Додатково можна мати:

* `1280x720`
* `2560x1440`
* `3840x2160`
* custom width / height

### 14.4. Export settings

Потрібно підтримати:

* width;
* height;
* fps;
* start frame;
* end frame;
* output folder;
* file prefix;
* overwrite policy;
* pixel ratio / supersampling.

---

## 15. Структура інтерфейсу

Застосунок має мати 4 вкладки:

1. `Текст і стиль`
2. `Анімація / Preview`
3. `Експорт`
4. `Проєкт / Пресети`

---

## 16. Вкладка 1: Текст і стиль

Призначення:

* робота з текстом;
* вибір display mode;
* вибір мови / профілю символів;
* налаштування кольору і layout;
* статичне preview.

Мінімальні блоки:

### 16.1. Charset / Languages

* список підтримуваних мов / профілів;
* toggle `Uppercase only`;
* toggle автозаміни unsupported characters;
* кнопка перегляду таблиці підтримуваних символів.

### 16.2. Display type

* `Segment`
* `Dot-matrix`

### 16.3. Text

* multi-line input;
* character counter;
* line counter;
* кнопки:
  * `Clear`
  * `Paste`

### 16.4. Style

* color mode:
  * single color
  * per-character color
* main color;
* glow color;
* glow intensity;
* inactive segment color;
* inactive segment visibility.

### 16.5. Layout

* cell scale;
* character spacing;
* line spacing;
* alignment;
* padding.

### 16.6. Static preview

* preview area;
* zoom;
* checkerboard mode;
* black background mode;
* custom background preview mode.

---

## 17. Вкладка 2: Анімація / Preview

Призначення:

* жива перевірка анімації;
* налаштування таймінгів;
* керування `seed`;
* playback.

Мінімальні блоки:

### 17.1. Animation type

* `Scramble basic`
* `Scramble with lock`
* optional:
  * `Reveal then scramble then lock`

### 17.2. Random source

* `Digits only`
* `Letters only`
* `Alphanumeric`
* `Current charset only`
* toggle `Allow invalid random chars`
* toggle `Allow empty cell`

### 17.3. Timing

* total duration;
* intro delay;
* outro hold;
* per-character delay;
* random switch rate;
* `Seed` field;
* button `Generate random seed`.

### 17.4. Reveal / Lock logic

* order;
* lock mode;
* toggle simultaneous final lock;
* toggle immediate punctuation lock.

### 17.5. Visual effects

* glow;
* flicker;
* noise;
* scanline toggle.

### 17.6. Playback

* Play
* Pause
* Stop
* Restart
* Loop
* timeline scrubber
* current frame indicator
* current time indicator

### 17.7. Preview area

* animated preview;
* zoom;
* safe area toggle;
* checkerboard toggle;
* optional segment outlines toggle.

---

## 18. Вкладка 3: Експорт

Призначення:

* рендер кадрів;
* формування команди `ffmpeg`;
* контроль параметрів export.

Мінімальні блоки:

### 18.1. Output

* export type:
  * `PNG sequence`
* output folder;
* file prefix;
* overwrite policy.

### 18.2. Render size

* width;
* height;
* presets:
  * `1920x1080`
  * `1080x1920`
  * optional additional presets
* supersampling.

### 18.3. Frame settings

* fps;
* start frame;
* end frame;
* preview region / full canvas.

### 18.4. Background

* background mode;
* color / gradient settings;
* image picker;
* video picker;
* fit mode.

### 18.5. FFmpeg

* auto-detected path або ручний path;
* згенерована команда `ffmpeg`;
* кнопка copy command.

### 18.6. Render control

* render button;
* progress bar;
* ETA;
* log window;
* cancel button;
* open output folder.

---

## 19. Вкладка 4: Проєкт / Пресети

Призначення:

* збереження стану;
* повторне використання preset-ів;
* швидке відкриття проєктів.

Мінімальні блоки:

### 19.1. Project

* New
* Open
* Save
* Save as
* Recent projects

### 19.2. Presets

* Save style preset
* Save animation preset
* Save export preset
* Apply preset
* Duplicate preset
* Delete preset

### 19.3. Metadata

* project name;
* notes;
* tags;
* created at;
* updated at.

---

## 20. Формат збереження

Для `v1.0`:

* project file: `JSON`
* presets: `JSON`

Проєкт повинен зберігати:

* текст;
* перелік активних мов / charset profile;
* display mode;
* style settings;
* layout settings;
* animation settings;
* background settings;
* export settings;
* `seed`;
* metadata.

---

## 21. Нефункціональні вимоги

### 21.1. Продуктивність

* preview у робочій роздільності має бути достатньо плавним для налаштування анімації;
* export може бути повільнішим за preview;
* export не повинен залежати від FPS preview.

### 21.2. Відтворюваність

* однаковий проєкт + однаковий `seed` = однаковий результат;
* preview і export не повинні відрізнятися за логікою анімації.

### 21.3. Якість

* transparent PNG має мати коректний alpha;
* glow має коректно працювати як на прозорому, так і на непрозорому фоні;
* текст має бути читабельний на image/video background.

### 21.4. Надійність

* застосунок повинен перевіряти наявність `ffmpeg`, якщо користувач хоче сформувати або виконувати video workflow;
* помилки рендера мають потрапляти в лог;
* export має коректно скасовуватися без пошкодження вже готових кадрів.

---

## 22. Acceptance criteria для v1.0

`v1.0` вважається прийнятою, якщо:

* користувач може створити новий проєкт;
* може ввести текст у кілька рядків;
* може вибрати `Segment` або `Dot-matrix`;
* може працювати принаймні з English, German, Ukrainian і Russian наборами символів;
* може застосувати fixed-width layout;
* може змінити колір, glow і базові animation settings;
* може задати або згенерувати `seed`;
* може побачити preview;
* може використати image або video background;
* може відрендерити `PNG sequence`;
* може отримати готову команду `ffmpeg` для складання відео;
* може зберегти проєкт і відкрити його без втрати налаштувань.

---

## 23. Межі v1.0

У `v1.0` не обов'язково включати:

* custom glyph editor;
* складний timeline editor;
* keyframes;
* MOV ProRes 4444;
* batch rendering;
* browser version;
* audio handling для background video;
* advanced color grading background video.

Ці функції не є помилкою відсутності у `v1.0`, якщо базовий workflow працює стабільно.
