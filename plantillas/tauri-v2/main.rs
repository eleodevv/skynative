// Punto de entrada de la aplicación de escritorio generada por Skynative.
//
// No se registra ningún comando nativo ni plugin: la ventana carga únicamente
// la URL remota indicada en tauri.conf.json. De este modo el contenido web no
// recibe ninguna capacidad privilegiada del sistema.

// En Windows se evita abrir una consola adicional en las compilaciones de
// publicación. Esta directiva debe permanecer en la primera línea útil.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("error al ejecutar la aplicación Tauri");
}
