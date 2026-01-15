export class EnvLoader {
  /**
   * Get string value or throw error if missing
   */
  public static getString(key: string): string {
    const value = process.env[key];
    if (value === undefined || value === '') {
      throw new Error(`Missing required environment variable: ${key}`);
    }
    return value;
  }

  /**
   * Get string value or default
   */
  public static getStringOptional(key: string, defaultValue: string = ''): string {
    const value = process.env[key];
    return (value !== undefined && value !== '') ? value : defaultValue;
  }

  /**
   * Get number value or throw/default
   */
  public static getNumber(key: string, defaultValue?: number): number {
    const value = process.env[key];
    if (value === undefined || value === '') {
      if (defaultValue !== undefined) return defaultValue;
      throw new Error(`Missing required environment variable: ${key}`);
    }
    const num = Number(value);
    if (isNaN(num)) {
      throw new TypeError(`Environment variable ${key} is not a valid number`);
    }
    return num;
  }

  /**
   * Get boolean value (true/false, 1/0, yes/no)
   */
  public static getBoolean(key: string, defaultValue: boolean = false): boolean {
    const value = process.env[key];
    if (value === undefined || value === '') {
      return defaultValue;
    }
    const normalized = value.toLowerCase().trim();
    return ['true', '1', 'yes', 'on'].includes(normalized);
  }
}
