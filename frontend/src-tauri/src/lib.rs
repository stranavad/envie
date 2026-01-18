// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use tauri::{Manager, TitleBarStyle, WebviewUrl, WebviewWindowBuilder};
use walkdir::WalkDir;

#[tauri::command]
fn scan_for_configs() -> Vec<String> {
    let mut files = Vec::new();
    let path = "/Users/davidstranava/programming";
    for entry in WalkDir::new(path)
        .into_iter()
        .filter_entry(|e| {
            !e.file_name()
                .to_string_lossy()
                .eq_ignore_ascii_case("node_modules")
        })
        .filter_map(|e| e.ok())
    {
        let file_name = entry.file_name().to_string_lossy();
        if file_name == ".env" || file_name == "config.local.yaml" {
            files.push(entry.path().to_string_lossy().to_string());
        }
    }
    files
}

#[tauri::command]
fn read_config_file(path: String) -> Result<String, String> {
    std::fs::read_to_string(path).map_err(|e| e.to_string())
}

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[tauri::command]
fn nuke_vault(app: tauri::AppHandle, user_id: String) -> Result<(), String> {
    let local_data_dir = app.path().app_local_data_dir().map_err(|e| e.to_string())?;

    // We do NOT delete salt.txt anymore because it might be shared (or we assume single user per OS account?)
    // If we want true multi-user, we should keep salt. But if loop fails, maybe we need to?
    // Let's assume for now we only delete the specific vault file.

    // Vault filename convention: "vault_<user_id>.hold"
    let vault_name = format!("vault_{}.hold", user_id);
    let vault_path = local_data_dir.join(&vault_name);

    if vault_path.exists() {
         std::fs::remove_file(&vault_path).map_err(|e| e.to_string())?;
    }

    // Also check standard filenames if user_id is empty or legacy?
    if user_id == "legacy" || user_id.is_empty() {
         let legacy_path = local_data_dir.join("vault.hold");
         if legacy_path.exists() { std::fs::remove_file(&legacy_path).map_err(|e| e.to_string())?; }

         let snapshot_path = local_data_dir.join("snapshot.hold");
         if snapshot_path.exists() { std::fs::remove_file(&snapshot_path).map_err(|e| e.to_string())?; }
    }

    Ok(())
}

#[tauri::command]
fn check_vault_exists(app: tauri::AppHandle, user_id: String) -> Result<bool, String> {
    let local_data_dir = app.path().app_local_data_dir().map_err(|e| e.to_string())?;

    let vault_name = format!("vault_{}.hold", user_id);
    let vault_path = local_data_dir.join(&vault_name);

    // Legacy fallback check?
    if !vault_path.exists() && (user_id == "legacy" || user_id.is_empty()) {
        let legacy = local_data_dir.join("vault.hold");
        let snapshot = local_data_dir.join("snapshot.hold");
        return Ok(legacy.exists() || snapshot.exists());
    }

    Ok(vault_path.exists())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .setup(|app| {
            let mut win_builder = WebviewWindowBuilder::new(app, "main", WebviewUrl::default())
                .title("Envie")
                .inner_size(1200.0, 800.0);

            #[cfg(target_os = "macos")]
            {
                win_builder = win_builder.title_bar_style(TitleBarStyle::Transparent);
            }

            let window = win_builder.build().unwrap();

            #[cfg(target_os = "macos")]
            {
                use objc2::rc::Retained;
                use objc2_app_kit::NSColor;
                use objc2_app_kit::NSWindow;
                use objc2_foundation::MainThreadMarker;

                // Recover the raw pointer from Tauri
                let ns_window_ptr = window.ns_window().unwrap();

                // Use unsafe new_unchecked as per compiler suggestion/docs for 0.2.2 if new() is missing/deprecated
                // Or maybe just assume we are on main thread (Tauri setup is).
                let _mtm = unsafe { MainThreadMarker::new_unchecked() };

                // Safety: trust the pointer
                let ns_window: Retained<NSWindow> = unsafe {
                    Retained::from_raw(ns_window_ptr as *mut NSWindow).expect("Failed to retain NSWindow")
                };

                let bg_color = unsafe {
                     NSColor::colorWithRed_green_blue_alpha(0.0, 0.0, 0.0, 1.0)
                };
                ns_window.setBackgroundColor(Some(&bg_color));
            }

            let salt_path = app
                .path()
                .app_local_data_dir()
                .expect("could not resolve app local data path")
                .join("salt.txt");
            app.handle().plugin(tauri_plugin_stronghold::Builder::with_argon2(&salt_path).build())?;

            Ok(())
        })
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_dialog::init())
        .invoke_handler(tauri::generate_handler![
            greet,
            scan_for_configs,
            read_config_file,
            nuke_vault,
            check_vault_exists
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
