import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from 'vscode-languageclient/node';

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext) {
  const binaryPath = resolveBinaryPath(context);
  if (!binaryPath) {
    vscode.window.showErrorMessage(
      'deoxy: binary not found. Install deoxy or set "deoxy.binaryPath" in settings.'
    );
    return;
  }

  const serverOptions: ServerOptions = {
    command: binaryPath,
    args: ['serve'],
    transport: TransportKind.stdio,
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: 'file', language: 'go' },
      { scheme: 'file', language: 'python' },
      { scheme: 'file', language: 'c' },
      { scheme: 'file', language: 'cpp' },
      { scheme: 'file', language: 'rust' },
    ],
  };

  client = new LanguageClient('deoxy', 'deoxy LSP Server', serverOptions, clientOptions);
  context.subscriptions.push(client.start());

  context.subscriptions.push(
    vscode.commands.registerCommand('deoxy.generateDoc', () => {
      vscode.commands.executeCommand('editor.action.codeAction', {
        kind: 'source',
        apply: 'ifSingle',
      }).then((result: any) => {
        if (!result) {
          vscode.window.showInformationMessage('No symbol found at cursor position');
        }
      });
    })
  );
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) return undefined;
  return client.stop();
}

function resolveBinaryPath(context: vscode.ExtensionContext): string | undefined {
  const configured = vscode.workspace.getConfiguration('deoxy').get<string>('binaryPath');
  if (configured) {
    if (isExecutable(configured)) return configured;
    vscode.window.showWarningMessage(
      `deoxy: configured binary not found or not executable: "${configured}". Falling back.`
    );
  }

  const bundled = context.asAbsolutePath(path.join('bin', platformBinaryName()));
  if (isExecutable(bundled)) return bundled;

  const fromPath = which('deoxy');
  if (fromPath) return fromPath;

  return undefined;
}

function platformBinaryName(): string {
  const platform = process.platform;
  const arch = process.arch;

  const goArch = arch === 'x64' ? 'amd64' : arch === 'arm64' ? 'arm64' : undefined;
  if (!goArch) return 'deoxy';

  switch (platform) {
    case 'linux':
      return `deoxy-linux-${goArch}`;
    case 'darwin':
      return `deoxy-darwin-${goArch}`;
    case 'win32':
      return `deoxy-windows-${goArch}.exe`;
    default:
      return 'deoxy';
  }
}

function isExecutable(filePath: string): boolean {
  try {
    const stat = fs.statSync(filePath);
    if (!stat.isFile()) return false;
    if (process.platform === 'win32') {
      const ext = path.extname(filePath).toLowerCase();
      return ext === '.exe' || ext === '.bat' || ext === '.cmd' || ext === '.com';
    }
    return (stat.mode & 0o111) !== 0;
  } catch {
    return false;
  }
}

function which(bin: string): string | undefined {
  const paths = process.env.PATH?.split(path.delimiter) ?? [];
  const isWin = process.platform === 'win32';
  const candidates = isWin
    ? [bin, `${bin}.exe`, `${bin}.bat`, `${bin}.cmd`]
    : [bin];

  for (const dir of paths) {
    for (const name of candidates) {
      const full = path.join(dir, name);
      if (isExecutable(full)) return full;
    }
  }
  return undefined;
}
