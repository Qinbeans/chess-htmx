import { spawnSync } from 'child_process';

spawnSync("dist/chess-htmx", [], { stdio: "inherit" });