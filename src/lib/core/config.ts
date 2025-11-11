/**
 * Frontend Configuration Service
 * Loads YAML-based configuration with environment overrides
 * Singleton pattern for global access
 */

import { load } from 'js-yaml';

// Core configuration interfaces with strict typing
export interface APISettings {
  readonly baseURL: string;
  readonly tokenKey: string;
  readonly timeout: number;
  readonly retryAttempts: number;
}

export interface UISettings {
  readonly theme: 'dark' | 'light';
  readonly language: 'fa' | 'en';
  readonly animationDuration: number;
  readonly globeAnimations: boolean;
}

export interface LoggingSettings {
  readonly level: 'debug' | 'info' | 'warn' | 'error';
  readonly enableRemote: boolean;
  readonly bufferSize: number;
}

export interface ExternalServices {
  readonly geoIPPrimary: string;
  readonly geoIPFallback: string;
  readonly pingNodePrimary: string;
  readonly pingNodeFallback: string;
}

// Main configuration structure
export interface Config {
  readonly api: APISettings;
  readonly ui: UISettings;
  readonly logging: LoggingSettings;
  readonly external: ExternalServices;
  readonly version: string;
  readonly environment: string;
}

// Text content structure for Persian UI
export interface TextContent {
  readonly auth: Record<string, string>;
  readonly dashboard: Record<string, string>;
  readonly devices: Record<string, string>;
  readonly tunnels: Record<string, string>;
  readonly operations: Record<string, string>;
  readonly common: Record<string, string>;
  readonly errors: Record<string, string>;
}

// Configuration loader class
class ConfigService {
  private static instance: ConfigService | null = null;
  private config: Config | null = null;
  private texts: TextContent | null = null;

  private constructor() {}

  // Singleton accessor
  public static getInstance(): ConfigService {
    if (!ConfigService.instance) {
      ConfigService.instance = new ConfigService();
    }
    return ConfigService.instance;
  }

  // Load configuration from YAML files
  public async loadConfig(environment: string = 'development'): Promise<Config> {
    try {
      // Load base configuration
      const baseConfigResponse = await fetch('/data/config/frontend.yml');
      if (!baseConfigResponse.ok) {
        throw new Error(`Failed to load base config: ${baseConfigResponse.statusText}`);
      }
      const baseConfigText = await baseConfigResponse.text();
      const baseConfig = load(baseConfigText) as Partial<Config>;

      // Load environment-specific overrides
      let envConfig: Partial<Config> = {};
      try {
        const envConfigResponse = await fetch(`/data/config/frontend.${environment}.yml`);
        if (envConfigResponse.ok) {
          const envConfigText = await envConfigResponse.text();
          envConfig = load(envConfigText) as Partial<Config>;
        }
      } catch {
        // Environment config is optional, continue without it
      }

      // Merge configurations with environment overrides
      this.config = this.mergeConfigs(this.getDefaultConfig(), baseConfig, envConfig);
      return this.config;
    } catch (error) {
      console.error('Config loading failed, using defaults:', error);
      this.config = this.getDefaultConfig();
      return this.config;
    }
  }

  // Load Persian UI texts from YAML
  public async loadTexts(): Promise<TextContent> {
    try {
      const textsResponse = await fetch('/data/texts/messages_fa.yml');
      if (!textsResponse.ok) {
        throw new Error(`Failed to load texts: ${textsResponse.statusText}`);
      }
      const textsYaml = await textsResponse.text();
      this.texts = load(textsYaml) as TextContent;
      return this.texts;
    } catch (error) {
      console.error('Text loading failed:', error);
      this.texts = this.getDefaultTexts();
      return this.texts;
    }
  }

  // Get current configuration (throws if not loaded)
  public getConfig(): Config {
    if (!this.config) {
      throw new Error('Configuration not loaded. Call loadConfig() first.');
    }
    return this.config;
  }

  // Get current texts (throws if not loaded)
  public getTexts(): TextContent {
    if (!this.texts) {
      throw new Error('Texts not loaded. Call loadTexts() first.');
    }
    return this.texts;
  }

  // Merge multiple configuration objects
  private mergeConfigs(...configs: Partial<Config>[]): Config {
    const merged = {} as Config;
    configs.forEach(config => {
      if (config) {
        Object.assign(merged, config);
      }
    });
    return merged;
  }

  // Default configuration fallback
  private getDefaultConfig(): Config {
    return {
      api: {
        baseURL: 'http://localhost:8080',
        tokenKey: 'vpn_auth_token',
        timeout: 5000,
        retryAttempts: 3
      },
      ui: {
        theme: 'dark',
        language: 'fa',
        animationDuration: 300,
        globeAnimations: true
      },
      logging: {
        level: 'info',
        enableRemote: false,
        bufferSize: 100
      },
      external: {
        geoIPPrimary: 'https://ipapi.ipspeed.info',
        geoIPFallback: 'https://findip.net',
        pingNodePrimary: 'ir1.node.check-host.net',
        pingNodeFallback: 'ir6.node.check-host.net'
      },
      version: '1.0.0',
      environment: 'development'
    };
  }

  // Default Persian texts fallback
  private getDefaultTexts(): TextContent {
    return {
      auth: { login: 'ورود', logout: 'خروج' },
      dashboard: { title: 'داشبورد' },
      devices: { title: 'دستگاه‌ها' },
      tunnels: { title: 'تونل‌ها' },
      operations: { title: 'عملیات' },
      common: { save: 'ذخیره', cancel: 'لغو', loading: 'در حال بارگیری...' },
      errors: { network: 'خطای شبکه', unknown: 'خطای نامشخص' }
    };
  }
}

// Export singleton instance
export const configService = ConfigService.getInstance();

// Export convenience functions
export const loadConfig = (env?: string) => configService.loadConfig(env);
export const loadTexts = () => configService.loadTexts();
export const getConfig = () => configService.getConfig();
export const getTexts = () => configService.getTexts();
