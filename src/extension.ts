/* // The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('Congratulations, your extension "go-checker-2" is now active!');

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with registerCommand
	// The commandId parameter must match the command field in package.json
	const disposable = vscode.commands.registerCommand('go-checker-2.helloWorld', () => {
		// The code you place here will be executed every time your command is executed
		// Display a message box to the user
		vscode.window.showInformationMessage('Hello World from go-checker-2!');
	});

	context.subscriptions.push(disposable);
}

// This method is called when your extension is deactivated
export function deactivate() {} */

// ================================================================ 2 ================================================================
/* import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as path from 'path';

interface VarInfo {
    name: string;
    line: number;
    kind: string;
}

export function activate(context: vscode.ExtensionContext) {
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('goChecker');
    context.subscriptions.push(diagnosticCollection);

    vscode.workspace.onDidSaveTextDocument((document) => {
        if (document.languageId !== 'go') {return;}

        const filePath = document.fileName;
        const extensionPath = context.extensionPath;
        const checkerPath = path.join(extensionPath, 'go-checker'); // compiled go-checker binary

        console.log("🛠️ Running Go Checker on:", filePath);

        const process = cp.spawn(checkerPath, [filePath]);

        let output = '';
        let error = '';

        process.stdout.on('data', (data) => {
            output += data.toString();
        });

        process.stderr.on('data', (data) => {
            error += data.toString();
        });

        process.on('close', () => {
            if (error) {
                vscode.window.showErrorMessage("Go Checker error: " + error);
                return;
            }

            let varInfos: VarInfo[] = [];
            try {
                varInfos = JSON.parse(output);
            } catch (err) {
                vscode.window.showErrorMessage("Failed to parse go-checker output");
                return;
            }

            const diagnostics: vscode.Diagnostic[] = [];

            varInfos.forEach(info => {
                const { name, line } = info;
                const lineIndex = line - 1;
                const lineText = document.lineAt(lineIndex).text;
                const index = lineText.indexOf(name);

				console.log('Name :', name);

                if (index === -1) {return;}

                const range = new vscode.Range(lineIndex, index, lineIndex, index + name.length);
                let message = '';

                if (!name.startsWith('l')) {
                    message = `Variable "${name}" should start with 'l'`;
                }
                if (name.endsWith('Arr') && !name.includes('[]')) {
                    message = `Variable "${name}" ends with 'Arr' but is not an array`;
                }
                if (name.endsWith('Map') && !name.includes('map')) {
                    message = `Variable "${name}" ends with 'Map' but is not a map`;
                }

                if (message) {
                    diagnostics.push(new vscode.Diagnostic(range, message, vscode.DiagnosticSeverity.Warning));
                }
            });

            diagnosticCollection.set(document.uri, diagnostics);
        });
    });
}

export function deactivate() {} */


// ================================================================ 3 ================================================================

import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as path from 'path';

export function activate(context: vscode.ExtensionContext) {
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('goChecker');
    context.subscriptions.push(diagnosticCollection);

    const runGoChecker = (document:vscode.TextDocument) => {
        if (document.languageId !== 'go') {
            return;
        }

        const filePath = document.fileName;
        // const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || '';
        // const goCheckerPath = path.join(context.extensionPath, 'go-checker'); // adjust this path if needed
        const goCheckerPath = context.asAbsolutePath('go-checker');

        console.log(`[Go Checker] Running for file: ${filePath}`);
        // const command = `${goCheckerPath} "${filePath}"`;
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
                        new vscode.Position(result.line - 1, result.column + result.name.length -1)
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

export function deactivate() {}
