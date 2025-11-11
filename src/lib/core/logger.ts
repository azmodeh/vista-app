/**
 * Client-Side Structured Logger Service
 * Enforces structured logging with strict API compliance
 * Replaces all console.log usage across the application
 */

import type { LoggingSettings } from './config';

// Log levels enum for type safety
export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  FATAL = 4
}

// Log entry structure for consistency
export interface LogEntry {
  readonly timestamp: string;
  readonly level: LogLevel;
  readonly message: string;
  readonly context?: Record<string, unknown>;
  readonly error?: Error;
  readonly component?: string;
}

// Logger service interface for contract compliance
export interface LoggerService {
  init(settings: LoggingSettings): void;
  debug(message: string, context?: Record<string, unknown>): void;
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, error?: Error, context?: Record<string, unknown>): void;
  fatal(message: string, error?: Error, context?: Record<string, unknown>): void;
  setComponent(componentName: string): LoggerService;
  flush(): Promise<void>;
}

// Core logger implementation
class BrowserLoggerService implements LoggerService {
  private static instance: BrowserLoggerService | null = null;
  private settings: LoggingSettings | null = null;
  private logBuffer: LogEntry[] = [];
  private currentLevel: LogLevel = LogLevel.INFO;
  private component?: string;

  private constructor() {}

  // Singleton accessor
  public static getInstance(): BrowserLoggerService {
    if (!BrowserLoggerService.instance) {
      BrowserLoggerService.instance = new BrowserLoggerService();
    }
    return BrowserLoggerService.instance;
  }

  // Initialize logger with configuration
  public init(settings: LoggingSettings): void {
    this.settings = settings;
    this.currentLevel = this.parseLogLevel(settings.level);
    
    // Setup periodic buffer flush if remote logging enabled
    if (settings.enableRemote) {
      setInterval(() => {
        this.flushToRemote().catch(() => {
          // Silent fail for remote logging to prevent recursion
        });
      }, 10000); // Flush every 10 seconds
    }
  }

  // Debug level logging
  public debug(message: string, context?: Record<string, unknown>): void {
    if (this.shouldLog(LogLevel.DEBUG)) {
      this.writeLog(LogLevel.DEBUG, message, context);
    }
  }

  // Info level logging
  public info(message: string, context?: Record<string, unknown>): void {
    if (this.shouldLog(LogLevel.INFO)) {
      this.writeLog(LogLevel.INFO, message, context);
    }
  }

  // Warning level logging
  public warn(message: string, context?: Record<string, unknown>): void {
    if (this.shouldLog(LogLevel.WARN)) {
      this.writeLog(LogLevel.WARN, message, context);
    }
  }

  // Error level logging
  public error(message: string, error?: Error, context?: Record<string, unknown>): void {
    if (this.shouldLog(LogLevel.ERROR)) {
      this.writeLog(LogLevel.ERROR, message, context, error);
    }
  }

  // Fatal level logging (always logged)
  public fatal(message: string, error?: Error, context?: Record<string, unknown>): void {
    this.writeLog(LogLevel.FATAL, message, context, error);
  }

  // Create contextual logger instance
  public setComponent(componentName: string): LoggerService {
    const contextual = Object.create(this) as BrowserLoggerService;
    contextual.component = componentName;
    return contextual;
  }

  // Flush pending logs to remote endpoint
  public async flush(): Promise<void> {
    if (this.settings?.enableRemote && this.logBuffer.length > 0) {
      await this.flushToRemote();
    }
  }

  // Check if log level should be written
  private shouldLog(level: LogLevel): boolean {
    return level >= this.currentLevel;
  }

  // Core log writing method
  private writeLog(
    level: LogLevel,
    message: string,
    context?: Record<string, unknown>,
    error?: Error
  ): void {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message: this.component ? `[${this.component}] ${message}` : message,
      context,
      error,
      component: this.component
    };

    // Add to buffer for potential remote logging
    this.addToBuffer(entry);

    // Output to browser console with appropriate method
    this.outputToBrowser(entry);
  }

  // Add log entry to internal buffer
  private addToBuffer(entry: LogEntry): void {
    if (!this.settings) return;

    this.logBuffer.push(entry);
    
    // Maintain buffer size limit
    if (this.logBuffer.length > this.settings.bufferSize) {
      this.logBuffer.shift(); // Remove oldest entry
    }
  }

  // Output log entry to browser console
  private outputToBrowser(entry: LogEntry): void {
    const formattedMessage = this.formatMessage(entry);
    const consoleArgs: any[] = [formattedMessage];
    
    if (entry.context) {
      consoleArgs.push(entry.context);
    }
    
    if (entry.error) {
      consoleArgs.push(entry.error);
    }

    // Route to appropriate console method
    switch (entry.level) {
      case LogLevel.DEBUG:
        console.debug(...consoleArgs);
        break;
      case LogLevel.INFO:
        console.info(...consoleArgs);
        break;
      case LogLevel.WARN:
        console.warn(...consoleArgs);
        break;
      case LogLevel.ERROR:
      case LogLevel.FATAL:
        console.error(...consoleArgs);
        break;
      default:
        console.log(...consoleArgs);
    }
  }

  // Format log message for console output
  private formatMessage(entry: LogEntry): string {
    const levelName = LogLevel[entry.level];
    const timestamp = entry.timestamp.substring(11, 23); // HH:mm:ss.SSS
    return `${timestamp} [${levelName}] ${entry.message}`;
  }

  // Parse string log level to enum
  private parseLogLevel(level: string): LogLevel {
    switch (level.toLowerCase()) {
      case 'debug': return LogLevel.DEBUG;
      case 'info': return LogLevel.INFO;
      case 'warn': return LogLevel.WARN;
      case 'error': return LogLevel.ERROR;
      case 'fatal': return LogLevel.FATAL;
      default: return LogLevel.INFO;
    }
  }

  // Flush logs to remote logging endpoint
  private async flushToRemote(): Promise<void> {
    if (this.logBuffer.length === 0) return;

    try {
      // Clone and clear buffer atomically
      const logsToSend = [...this.logBuffer];
      this.logBuffer.length = 0;

      // Send to remote logging endpoint
      await fetch('/api/v1/logs/client', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          entries: logsToSend,
          userAgent: navigator.userAgent,
          url: window.location.href
        })
      });
    } catch (error) {
      // Restore logs to buffer on failure, prepend to maintain order
      this.logBuffer.unshift(...this.logBuffer);
    }
  }
}

// Export singleton instance
export const logger = BrowserLoggerService.getInstance();

// Export convenience functions for global use
export const initLogger = (settings: LoggingSettings) => logger.init(settings);
export const debug = (msg: string, ctx?: Record<string, unknown>) => logger.debug(msg, ctx);
export const info = (msg: string, ctx?: Record<string, unknown>) => logger.info(msg, ctx);
export const warn = (msg: string, ctx?: Record<string, unknown>) => logger.warn(msg, ctx);
export const error = (msg: string, err?: Error, ctx?: Record<string, unknown>) => 
  logger.error(msg, err, ctx);
export const fatal = (msg: string, err?: Error, ctx?: Record<string, unknown>) => 
  logger.fatal(msg, err, ctx);

// Export component-specific logger factory
export const getLogger = (component: string) => logger.setComponent(component);
