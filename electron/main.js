const { app, BrowserWindow, shell } = require('electron');
const path = require('path');

const PROTOCOL = process.env.ELECTRON_PROTOCOL || 'bgl';
const START_URL =
  process.env.ELECTRON_START_URL ||
  `file://${path.join(__dirname, '../frontend/dist/index.html')}`;
const INITIAL_ROUTE = process.env.ELECTRON_INITIAL_ROUTE || '/m?ui=mobile';

let mainWindow;
let pendingDeepLink = null;

const buildAppUrl = (route) => {
  const baseUrl = START_URL.split('#')[0];
  const normalized = route.startsWith('/') ? route : `/${route}`;
  return `${baseUrl}#${normalized}`;
};

const handleDeepLink = (url) => {
  try {
    const parsed = new URL(url);
    const route = `${parsed.pathname || '/'}${parsed.search || ''}`;
    if (mainWindow) {
      mainWindow.loadURL(buildAppUrl(route));
      mainWindow.focus();
    } else {
      pendingDeepLink = route;
    }
  } catch (error) {
    console.error('Failed to handle deep link:', error);
  }
};

const createWindow = () => {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    backgroundColor: '#0b0b0b',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  const initialUrl = pendingDeepLink || INITIAL_ROUTE;
  mainWindow.loadURL(buildAppUrl(initialUrl));

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
};

const ensureSingleInstance = () => {
  const gotLock = app.requestSingleInstanceLock();
  if (!gotLock) {
    app.quit();
    return;
  }

  app.on('second-instance', (_event, argv) => {
    const deepLinkArg = argv.find((arg) => arg.startsWith(`${PROTOCOL}://`));
    if (deepLinkArg) {
      handleDeepLink(deepLinkArg);
    }
    if (mainWindow) {
      if (mainWindow.isMinimized()) {
        mainWindow.restore();
      }
      mainWindow.focus();
    }
  });
};

app.on('open-url', (event, url) => {
  event.preventDefault();
  handleDeepLink(url);
});

ensureSingleInstance();

app.whenReady().then(() => {
  app.setAsDefaultProtocolClient(PROTOCOL);
  createWindow();

  const deepLinkArg = process.argv.find((arg) => arg.startsWith(`${PROTOCOL}://`));
  if (deepLinkArg) {
    handleDeepLink(deepLinkArg);
  }

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});
