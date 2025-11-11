/**
 * Main Application Initialization
 * Handles core service startup and configuration loading
 */

import { loadConfig, loadTexts } from './config';
import { initLogger, info, error } from './logger';

// Global application state
let isInitialized = false;

/**
 * Initialize the application with all core services
 * Must be called before any other application logic
 */
export async function initApp(): Promise<void> {
  if (isInitialized) {
    return;
  }

  try {
    // Load configuration first
    const config = await loadConfig();
    
    // Initialize logger with config
    initLogger(config.logging);
    info('Application initialization started', { version: config.version });

    // Load UI texts
    await loadTexts();
    info('UI texts loaded successfully');

    // Mark as initialized
    isInitialized = true;
    info('Application initialization completed successfully');

  } catch (err) {
    console.error('Critical: Application initialization failed:', err);
    throw err;
  }
}

/**
 * Check if application is properly initialized
 */
export function isAppInitialized(): boolean {
  return isInitialized;
}
