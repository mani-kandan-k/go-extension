import * as vscode from 'vscode';
import * as cp from 'child_process';

export function activate(context: vscode.ExtensionContext) {
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('goChecker');
    context.subscriptions.push(diagnosticCollection);

    const runGoChecker = (document: vscode.TextDocument) => {
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

            let diagnostics: vscode.Diagnostic[] = [];
            const results = JSON.parse(stdout);
            try {
                for (const result of results) {
                    const range = new vscode.Range(
                        new vscode.Position(result.line - 1, result.column - 1),
                        new vscode.Position(result.line - 1, result.column + result.name.length - 1)
                    );
                    const diagnostic = new vscode.Diagnostic(
                        range,
                        result.message,
                        vscode.DiagnosticSeverity.Warning
                    );
                    diagnostics.push(diagnostic);
                }
            } catch (parseErr) {
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

export function deactivate() { }
