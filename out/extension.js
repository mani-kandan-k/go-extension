"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const cp = __importStar(require("child_process"));
function activate(context) {
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('goChecker');
    context.subscriptions.push(diagnosticCollection);
    const runGoChecker = (document) => {
        if (document.languageId !== 'go') {
            return;
        }
        const filePath = document.fileName;
        const goCheckerPath = context.asAbsolutePath('go-checker');
        console.log(`[Go Checker] Running for file: ${filePath}`);
        const command = `"${goCheckerPath}" "${filePath}"`;
        cp.exec(command, (err, stdout, stderr) => {
            if (err) {
                console.error(`[Go Checker Error] ${stderr}`);
                vscode.window.showErrorMessage(`Go Checker error: ${stderr}`);
                return;
            }
            let diagnostics = [];
            const results = JSON.parse(stdout);
            try {
                for (const result of results) {
                    const range = new vscode.Range(new vscode.Position(result.line - 1, result.column - 1), new vscode.Position(result.line - 1, result.column + result.name.length - 1));
                    const diagnostic = new vscode.Diagnostic(range, result.message, vscode.DiagnosticSeverity.Warning);
                    diagnostics.push(diagnostic);
                }
            }
            catch (parseErr) {
                console.error('[Go Checker] Failed to parse output', parseErr);
                console.log("results :", results);
                return;
            }
            diagnosticCollection.set(document.uri, diagnostics);
        });
    };
    vscode.workspace.onDidSaveTextDocument(runGoChecker);
    vscode.workspace.onDidOpenTextDocument(runGoChecker);
}
function deactivate() { }
//# sourceMappingURL=extension.js.map