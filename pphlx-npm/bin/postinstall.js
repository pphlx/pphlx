const fs = require('fs');
const path = require('path');

// During npm install, INIT_CWD points to the directory where the install command was run
const projectDir = process.env.INIT_CWD || path.resolve(__dirname, '../../../');
const packageJsonPath = path.join(projectDir, 'package.json');

if (fs.existsSync(packageJsonPath)) {
  try {
    const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    
    if (!packageJson.scripts) {
      packageJson.scripts = {};
    }
    
    let modified = false;
    
    if (!packageJson.scripts.build || !packageJson.scripts.build.includes("pphlx")) {
      packageJson.scripts.build = "pphlx";
      modified = true;
    }
    if (!packageJson.scripts.dev) {
      packageJson.scripts.dev = "pphlx dev";
      modified = true;
    }
    if (!packageJson.scripts.start) {
      packageJson.scripts.start = "pphlx dev";
      modified = true;
    }
    if (!packageJson.scripts.watch) {
      packageJson.scripts.watch = "pphlx watch";
      modified = true;
    }
    if (!packageJson.scripts.preview) {
      packageJson.scripts.preview = "pphlx preview";
      modified = true;
    }
    if (!packageJson.scripts.check) {
      packageJson.scripts.check = "pphlx check";
      modified = true;
    }
    
    if (modified) {
      fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2), 'utf8');
      console.log(`[pphlx] Automatically added "dev", "build", and "watch" scripts to your package.json!`);
    }
  } catch (err) {
    console.warn(`[pphlx] Failed to auto-configure package.json scripts: ${err.message}`);
  }
}
