# Go_REST_webui

Mikrousługa REST napisana w Go z wbudowanym (embed) interfejsem webowym w stylu
panelu/dashboardu. Całość — backend API, logika i frontend (HTML/CSS/JS) —
kompiluje się do jednej binarki `.exe`, bez zewnętrznych zależności
uruchomieniowych i bez frameworków webowych (tylko biblioteka standardowa `net/http`).

## Struktura projektu

```
Go_REST_webui\
├───.vscode\              - konfiguracja VS Code (taski budowania, launch, historia)
├───bin\                  - skompilowane binarki (Go_REST_webui.exe)
├───src\
│   ├───main.go           - punkt wejścia: routing HTTP, embed plików statycznych
│   ├───main_test.go      - testy
│   ├───Internal\
│   │   ├───api\
│   │   │   └───server.go     - handlery HTTP + middleware (logowanie, licznik requestów)
│   │   └───store\
│   │       └───store.go      - magazyn danych w pamięci (thread-safe, sync.RWMutex)
│   └───web\
│       └───static\
│           ├───index.html    - layout panelu (GUI)
│           ├───style.css     - ciemny motyw, karty, tabela
│           └───app.js        - komunikacja z API (fetch), auto-odświeżanie
├───go.mod                - moduł: Go_REST_webui
└───README.md
```

> Uwaga: `go.mod` znajduje się w katalogu głównym, a kod w `src\`, dlatego
> importy wewnętrzne wyglądają tak: `Go_REST_webui/src/Internal/api`,
> `Go_REST_webui/src/Internal/store`.

## Uruchomienie

Z poziomu katalogu głównego (`Go_REST_webui\`):

```powershell
go run .\src
```

lub z poziomu wtyczki na belce statusu  > Go RUN 
lub z poziomu MacroPAD I/II  przycisk zielony  >  (nr6)

Domyślnie serwer wystartuje na porcie `8080` → http://localhost:8080
Port można nadpisać zmienną środowiskową:

```powershell
$env:PORT="9000"; go run .\src
```



## Jak działa program

Aplikacja to klasyczna mikrousługa REST z warstwami:

1. **`main.go`** — buduje router (`http.ServeMux`, routing ścieżkowy z Go 1.22:
   `GET /api/items/{id}` itd.), owija go middleware'ami (logowanie, licznik
   requestów) i serwuje jednocześnie API oraz statyczny frontend z tego samego
   procesu na tym samym porcie.
2. **`Internal/store`** — prosty magazyn danych w pamięci (`map[string]Item`)
   zabezpieczony `sync.RWMutex`, czyli bezpieczny przy równoległych zapytaniach.
   Model `Item` jest celowo generyczny (ID, Name, Status, znaczniki czasu) —
   łatwo zamienić go na własną domenę (np. status buildów Go, stan MacroPAD-a)
   bez zmiany reszty architektury.
3. **`Internal/api`** — handlery HTTP tłumaczące żądania na operacje na `store`
   oraz zwracające JSON. Middleware liczy requesty i loguje każdą operację
   (metoda, ścieżka, czas trwania).
4. **`web/static`** — frontend wbudowany w binarkę przez `//go:embed web/static`
   w `main.go`. Dzięki temu gotowy `.exe` nie wymaga żadnych plików obok siebie —
   HTML/CSS/JS są zaszyte w kodzie wynikowym.

### Endpointy API

| Metoda | Ścieżka           | Opis                              |
|--------|--------------------|-----------------------------------|
| GET    | `/api/health`      | health check                      |
| GET    | `/api/status`      | uptime, liczba requestów, RAM     |
| GET    | `/api/items`       | lista elementów                   |
| POST   | `/api/items`       | dodanie elementu                  |
| GET    | `/api/items/{id}`  | pobranie pojedynczego elementu    |
| PUT    | `/api/items/{id}`  | edycja elementu                   |
| DELETE | `/api/items/{id}`  | usunięcie elementu                |

## GUI (panel webowy)

Interfejs to jednostronicowy panel w ciemnym motywie, dostępny pod adresem
serwera (`http://localhost:8080`), zbudowany z czystego HTML/CSS/JS (bez
frameworka frontendowego) i komunikujący się z API przez `fetch`.

Panel składa się z trzech sekcji:

- **Karty statusu** (góra) — cztery żywe wskaźniki: liczba requestów, liczba
  elementów w magazynie, liczba goroutines, zużycie pamięci (MB). Odświeżają
  się automatycznie co 3 sekundy przez `/api/status`. Kropka przy nazwie
  usługi w nagłówku świeci na zielono, gdy backend odpowiada, i na czerwono
  przy utracie połączenia.
- **Panel elementów** (środek) — tabela z danymi ze `store`: ID, nazwa, status
  (kolorowa etykieta: nowy / w toku / gotowe), data utworzenia. Formularz nad
  tabelą pozwala dodać nowy element (`POST /api/items`). Każdy wiersz ma dwa
  przyciski: `⟳` cyklicznie zmienia status elementu (nowy → w toku → gotowe →
  nowy…) przez `PUT`, `✕` usuwa element przez `DELETE`.
- **Log zdarzeń** (dół) — lokalna, przeglądarkowa lista ostatnich akcji
  wykonanych w panelu (dodanie/edycja/usunięcie elementu), z przyciskiem do
  wyczyszczenia. To log po stronie klienta (JS), niezależny od logów backendu
  w konsoli serwera.

Layout jest responsywny (siatka kart przechodzi z 4 na 2 kolumny na wąskich
ekranach) i nie wymaga żadnych zewnętrznych bibliotek CSS/JS — cały styl jest
w `style.css` na zmiennych CSS (`:root { --bg, --accent, ... }`), co ułatwia
szybkie przekolorowanie motywu.

## Rozbudowa

- Podmień `Item` w `store.go` na docelową domenę (np. `BuildJob`, `Device`,
  `Task`) — interfejs `Store` zostaje bez zmian.
- Magazyn w pamięci można zamienić na SQLite/BoltDB, zachowując te same
  metody (`List`, `Get`, `Create`, `Update`, `Delete`).
- Do wersji produkcyjnej warto dodać: graceful shutdown (`context` +
  `srv.Shutdown`), konfigurację przez plik/flagi zamiast zmiennych
  środowiskowych, CORS (jeśli GUI miałoby być hostowane osobno od API), oraz
  rozbudowę `main_test.go` o testy handlerów i store'a.
- Panel można rozszerzyć o SSE (`text/event-stream`) zamiast pollingu co 3s,
  co dałoby log na żywo bez odpytywania serwera w pętli.
