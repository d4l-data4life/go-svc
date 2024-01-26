const fs = require("fs");
const readdirRecursive = require("fs-readdir-recursive");
const path = require("path");
const importRegex = /"github\.com\/gesundheitscloud\/go-svc\/pkg\/(.+)"/g;

const targetFile = "dependencies.plantuml";
const RepoConfig = [
  { path: "../../research-pillars", name: "ResearchPillars" },
  { path: "../../data-dispatcher", name: "DataDispatcher" },
  { path: "../../data-receiver", name: "DataReceiver" },
];

class DependenciesGraph {
  packages = new Set();
  repoDependencies = RepoConfig.reduce((acc, cur) => {
    acc[cur.name] = new Set();
    return acc;
  }, {});
  packageDependencies = {};

  loadAndGetGoSVCImports(filePath) {
    const data = fs.readFileSync(filePath, "utf8");
    const matches = [...data.matchAll(importRegex)];
    return matches;
  }

  printSet(set, printFn) {
    const setLines = Array.from(set)
      .sort((a, b) => a.localeCompare(b))
      .map(printFn);
    if (setLines.length > 0) {
      setLines.push("");
    }
    return setLines;
  }

  getRepoDependencies() {
    for (const source of RepoConfig) {
      const rootPaths = ["pkg", "internal", "cmd"].map((el) =>
        path.join(source.path, el)
      );
      for (const rootPath of rootPaths) {
        const files = readdirRecursive(rootPath);
        for (const file of files) {
          if (file.endsWith(".go")) {
            const filePath = path.join(rootPath, file);
            const imports = this.loadAndGetGoSVCImports(filePath).map(
              (match) => match[1]
            );
            for (const el of imports) {
              this.packages.add(el);
              this.repoDependencies[source.name].add(el);
            }
          }
        }
      }
    }
  }

  getPackageDependencies() {
    const basePath = path.join(__dirname, "..", "pkg");
    const packages = fs.readdirSync(basePath);
    for (const pkg of packages) {
      this.packages.add(pkg);
      this.packageDependencies[pkg] = new Set();
      const files = readdirRecursive(path.join(basePath, pkg));
      for (const file of files) {
        if (file.endsWith(".go")) {
          const filePath = path.join(basePath, pkg, file);
          const imports = this.loadAndGetGoSVCImports(filePath).map(
            (match) => match[1]
          );
          for (const el of imports) {
            if (el.split("/")[0] != pkg) {
              this.packageDependencies[pkg].add(el);
            }
          }
        }
      }
    }
  }

  printPlantUML() {
    let lines = ["@startuml Dependencies", "", "together {"];
    lines.push(
      ...Object.keys(this.repoDependencies).map((el) => `class ${el}`)
    );
    lines.push("}", "");
    lines.push(...this.printSet(this.packages, (el) => `object ${el}`));

    for (const [name, deps] of Object.entries(this.repoDependencies)) {
      lines.push(...this.printSet(deps, (el) => `${name} --> ${el}`));
    }

    for (const [name, deps] of Object.entries(this.packageDependencies)) {
      lines.push(...this.printSet(deps, (el) => `${name} -[dotted]-|> ${el}`));
    }

    lines.push("@enduml", "");
    return lines.join("\n");
  }

  create() {
    try {
      this.getRepoDependencies();
      this.getPackageDependencies();

      const uml = this.printPlantUML();
      console.log(uml);
      fs.writeFileSync(targetFile, uml);
    } catch (error) {
      console.log(error.stack);
    }
  }
}

new DependenciesGraph().create();
