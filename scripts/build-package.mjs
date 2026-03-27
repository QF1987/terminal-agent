import { execSync } from "node:child_process";

const packageDir = process.argv[2];

if (!packageDir) {
  console.error("缺少包目录参数");
  process.exit(1);
}

try {
  execSync(`npx tsc -p ${packageDir}/tsconfig.json`, {
    cwd: new URL("..", import.meta.url).pathname,
    stdio: "inherit"
  });
} catch {
  process.exit(1);
}
