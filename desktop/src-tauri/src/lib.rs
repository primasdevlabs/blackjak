use std::fs;
use std::net::TcpListener;
use std::path::{Path, PathBuf};
use std::process::{Child, Command};
use std::sync::Mutex;
use tauri::{Manager, State};
use tauri_plugin_dialog::{DialogExt, FilePath};
use tauri_plugin_notification::NotificationExt;

struct AgentRuntime {
    child: Mutex<Option<Child>>,
    port: Mutex<u16>,
    workspace: Mutex<String>,
}

fn free_port() -> Result<u16, String> {
    let listener = TcpListener::bind("127.0.0.1:0").map_err(|e| e.to_string())?;
    let port = listener.local_addr().map_err(|e| e.to_string())?.port();
    Ok(port)
}

fn find_agent_binary(resource_dir: Option<&Path>) -> Result<PathBuf, String> {
    let mut candidates: Vec<PathBuf> = Vec::new();
    if let Some(dir) = resource_dir {
        candidates.push(dir.join("agent.exe"));
        candidates.push(dir.join("agent"));
        candidates.push(dir.join("bin").join("agent.exe"));
        candidates.push(dir.join("bin").join("agent"));
    }
    if let Ok(exe) = std::env::current_exe() {
        if let Some(parent) = exe.parent() {
            candidates.push(parent.join("agent.exe"));
            candidates.push(parent.join("agent"));
            candidates.push(parent.join("bin").join("agent.exe"));
            candidates.push(parent.join("bin").join("agent"));
            // Dev: desktop/src-tauri/target/... → repo root bin/
            if let Some(repo) = parent
                .ancestors()
                .find(|p| p.join("go.mod").exists() || p.join("bin").exists())
            {
                candidates.push(repo.join("bin").join("agent.exe"));
                candidates.push(repo.join("bin").join("agent"));
            }
        }
    }
    for rel in [
        "bin/agent.exe",
        "bin/agent",
        "../bin/agent.exe",
        "../bin/agent",
        "../../bin/agent.exe",
        "../../bin/agent",
    ] {
        candidates.push(PathBuf::from(rel));
    }
    candidates
        .into_iter()
        .find(|p| p.exists())
        .ok_or_else(|| "agent binary not found; run make build-backend".to_string())
}

fn workspace_config_path(app: &tauri::AppHandle) -> Option<PathBuf> {
    app.path().app_config_dir().ok().map(|d| d.join("workspace.txt"))
}

fn load_workspace(app: &tauri::AppHandle) -> Option<String> {
    let path = workspace_config_path(app)?;
    let data = fs::read_to_string(path).ok()?;
    let trimmed = data.trim().to_string();
    if trimmed.is_empty() {
        None
    } else {
        Some(trimmed)
    }
}

fn save_workspace(app: &tauri::AppHandle, workspace: &str) -> Result<(), String> {
    let path = workspace_config_path(app).ok_or_else(|| "no config dir".to_string())?;
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent).map_err(|e| e.to_string())?;
    }
    fs::write(path, workspace).map_err(|e| e.to_string())
}

fn spawn_agent(binary: &Path, port: u16, workspace: &str) -> Result<Child, String> {
    Command::new(binary)
        .args([
            "--server",
            "--port",
            &port.to_string(),
            "--workspace",
            workspace,
        ])
        .spawn()
        .map_err(|e| format!("failed to spawn agent: {e}"))
}

fn inject_port(app: &tauri::AppHandle, port: u16) {
    if let Some(win) = app.get_webview_window("main") {
        let script = format!(
            "window.__AGENT_PORT__ = {port}; window.__AGENT_HOST__ = '127.0.0.1'; try {{ localStorage.setItem('blackjak.agentPort', '{port}'); }} catch (e) {{}}"
        );
        let _ = win.eval(&script);
    }
}

#[tauri::command]
fn agent_port(state: State<AgentRuntime>) -> u16 {
    *state.port.lock().unwrap()
}

#[tauri::command]
fn agent_workspace(state: State<AgentRuntime>) -> String {
    state.workspace.lock().unwrap().clone()
}

#[tauri::command]
async fn open_settings_window(app: tauri::AppHandle) -> Result<(), String> {
    if let Some(w) = app.get_webview_window("settings") {
        w.show().map_err(|e| e.to_string())?;
        w.set_focus().map_err(|e| e.to_string())?;
        return Ok(());
    }
    tauri::WebviewWindowBuilder::new(
        &app,
        "settings",
        tauri::WebviewUrl::App("index.html#/settings".into()),
    )
    .title("BlackJak Settings")
    .inner_size(960.0, 720.0)
    .build()
    .map_err(|e| e.to_string())?;
    Ok(())
}

#[tauri::command]
async fn open_activity_window(app: tauri::AppHandle) -> Result<(), String> {
    if let Some(w) = app.get_webview_window("activity") {
        w.show().map_err(|e| e.to_string())?;
        w.set_focus().map_err(|e| e.to_string())?;
        return Ok(());
    }
    tauri::WebviewWindowBuilder::new(
        &app,
        "activity",
        tauri::WebviewUrl::App("index.html#/activity".into()),
    )
    .title("BlackJak Activity")
    .inner_size(800.0, 600.0)
    .build()
    .map_err(|e| e.to_string())?;
    Ok(())
}

#[tauri::command]
async fn notify(app: tauri::AppHandle, title: String, body: String) -> Result<(), String> {
    app.notification()
        .builder()
        .title(title)
        .body(body)
        .show()
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn pick_files(
    app: tauri::AppHandle,
    multiple: Option<bool>,
    title: Option<String>,
) -> Result<Vec<String>, String> {
    let multiple = multiple.unwrap_or(true);
    let mut builder = app.dialog().file();
    if let Some(t) = title {
        builder = builder.set_title(t);
    }
    let picked = if multiple {
        builder.blocking_pick_files()
    } else {
        builder.blocking_pick_file().map(|p| vec![p])
    };
    Ok(picked
        .unwrap_or_default()
        .into_iter()
        .map(|p: FilePath| p.to_string())
        .collect())
}

#[tauri::command]
async fn pick_workspace(app: tauri::AppHandle, state: State<'_, AgentRuntime>) -> Result<String, String> {
    let folder = app
        .dialog()
        .file()
        .set_title("Choose workspace folder")
        .blocking_pick_folder()
        .ok_or_else(|| "workspace selection cancelled".to_string())?;
    let workspace = folder.to_string();
    save_workspace(&app, &workspace)?;
    *state.workspace.lock().unwrap() = workspace.clone();

    // Restart agent against the new workspace on the same port.
    let port = *state.port.lock().unwrap();
    let resource = app.path().resource_dir().ok();
    let binary = find_agent_binary(resource.as_deref())?;
    {
        let mut child_slot = state.child.lock().unwrap();
        if let Some(mut child) = child_slot.take() {
            let _ = child.kill();
            let _ = child.wait();
        }
        let child = spawn_agent(&binary, port, &workspace)?;
        *child_slot = Some(child);
    }
    inject_port(&app, port);
    Ok(workspace)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_notification::init())
        .setup(|app| {
            let port = free_port().unwrap_or(47811);
            let mut workspace = load_workspace(app.handle()).unwrap_or_else(|| {
                std::env::current_dir()
                    .map(|p| p.to_string_lossy().to_string())
                    .unwrap_or_else(|_| ".".into())
            });

            // First launch with no saved workspace: offer a folder picker.
            if load_workspace(app.handle()).is_none() {
                if let Some(folder) = app
                    .dialog()
                    .file()
                    .set_title("Choose BlackJak workspace folder")
                    .blocking_pick_folder()
                {
                    workspace = folder.to_string();
                    let _ = save_workspace(app.handle(), &workspace);
                }
            }

            let resource = app.path().resource_dir().ok();
            let runtime = AgentRuntime {
                child: Mutex::new(None),
                port: Mutex::new(port),
                workspace: Mutex::new(workspace.clone()),
            };

            match find_agent_binary(resource.as_deref()) {
                Ok(binary) => match spawn_agent(&binary, port, &workspace) {
                    Ok(child) => {
                        *runtime.child.lock().unwrap() = Some(child);
                        let _ = app
                            .notification()
                            .builder()
                            .title("BlackJak")
                            .body(format!("Agent on 127.0.0.1:{port}"))
                            .show();
                    }
                    Err(err) => eprintln!("[BlackJak Desktop] {err}"),
                },
                Err(err) => eprintln!("[BlackJak Desktop] {err}"),
            }

            app.manage(runtime);
            inject_port(app.handle(), port);

            // Re-inject after the webview finishes loading the UI.
            let handle = app.handle().clone();
            std::thread::spawn(move || {
                std::thread::sleep(std::time::Duration::from_millis(800));
                inject_port(&handle, port);
            });

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            agent_port,
            agent_workspace,
            open_settings_window,
            open_activity_window,
            notify,
            pick_files,
            pick_workspace
        ])
        .run(tauri::generate_context!())
        .expect("error while running BlackJak desktop");
}
