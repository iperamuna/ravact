# FrankenPHP Classic Mode Enhancements TODO

## Phase 1: Model & Detection
- [x] Update `FrankenPHPClassicModel` struct with new configuration fields (threads, wait time, PHP INI settings).
- [x] Implement `identifyExistingSetup()` to detect if any FrankenPHP classic sites exist.
- [x] Update Site Commands to redirect to `FrankenPHPServicesScreen` if setup exists.

## Phase 2: Enhanced Form
- [x] Update `huh` form in `viewSiteSetup` to include:
    - [x] `num_threads` (suggestion: logic for threads * 2).
    - [x] `max_threads` (integer or "auto").
    - [x] `max_wait_time` (default 15s).
- [x] Add a new section/group for PHP INI settings with provided defaults.
- [x] Implement "Edit" mode to load existing file values.

## Phase 3: Configuration Templates & File Review
- [x] Update `Caddyfile` template for Socket/Port modes.
- [x] Update Systemd Service template with `ExecStartPre`, `ExecStartPost`, `Restart`, and PHP INI path.
- [x] Create `app-php.ini` template.
- [x] Implement `viewFileConfirmation` screen:
    - [x] List: Caddyfile, Service File, app-php.ini.
    - [x] Navigation: Up/Down.
    - [x] Action 'v': View generated content.
    - [x] Action 'e': Edit with nano/vi.

## Phase 4: Service Deployment & Verification
- [x] Implement deployment logic:
    - [x] Write files to `/etc/frankenphp/{id}/` and `/etc/systemd/system/`.
    - [x] Run `systemctl daemon-reload`.
    - [x] Run `systemctl enable --now {id}`.
- [x] Implement service verification and PHP INI load check.

## Phase 5: Nginx Integration
- [x] Detect Nginx configuration for the given domain.
- [x] Parse config to find PHP-FPM location block.
- [x] Suggest replacement with FrankenPHP proxy (Socket or Port).
- [x] Implement Nginx config review and reload logic.
- [x] Add fallback for missing Nginx config (create/manual edit).
