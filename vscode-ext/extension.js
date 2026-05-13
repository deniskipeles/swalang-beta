const path = require('path');
const fs = require('fs');
const vscode = require('vscode');
const { LanguageClient } = require('vscode-languageclient/node');

let client;
let outputChannel;

function activate(context) {
    outputChannel = vscode.window.createOutputChannel("Swalang LSP");
    outputChannel.appendLine('Starting Swalang LSP Extension...');

    // 1. Determine Binary Path
    let swalangPath = process.env.SWALANG_PATH;
    
    // Cloud IDE Fallback for Lightning.ai
    if (!swalangPath) {
        outputChannel.appendLine('SWALANG_PATH not found. Trying cloud IDE fallback...');
        const cloudFallback = '/teamspace/studios/this_studio/swalang-beta/builds/linux-x86_64/bin';
        if (fs.existsSync(cloudFallback)) {
            swalangPath = cloudFallback;
        }
    }

    if (!swalangPath) {
        vscode.window.showErrorMessage("Swalang LSP: SWALANG_PATH is not set and fallback failed.");
        outputChannel.appendLine("ERROR: Binary path not found.");
        return;
    }

    const isWindows = process.platform === 'win32';
    const binaryName = isWindows ? 'swalang-lsp.exe' : 'swalang-lsp';
    const serverPath = path.join(swalangPath, binaryName);

    outputChannel.appendLine(`Resolved LSP Binary: ${serverPath}`);

    if (!fs.existsSync(serverPath)) {
        vscode.window.showErrorMessage(`Swalang LSP: Binary not found at ${serverPath}`);
        outputChannel.appendLine("ERROR: Executable does not exist on disk.");
        return;
    }

    const serverOptions = {
        command: serverPath,
        args: []
    };

    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'swalang' }],
        outputChannel: outputChannel, // Pipe all extension logs here
    };

    client = new LanguageClient(
        'swalangLSP',
        'Swalang Language Server',
        serverOptions,
        clientOptions
    );

    client.start();
    outputChannel.appendLine('LSP Client started successfully.');
}

function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

module.exports = {
    activate,
    deactivate
};