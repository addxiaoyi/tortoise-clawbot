
export class StructureParser {
  public parseJSON<T = unknown>(input: string): T {
    try {
      return JSON.parse(input);
    } catch (e) {
      const jsonMatch = input.match(/```(?:json)?\s*([\s\S]*?)\s*```/);
      if (jsonMatch && jsonMatch[1]) {
        try {
          return JSON.parse(jsonMatch[1]);
        } catch (innerError) {
          const msg = innerError instanceof Error ? innerError.message : String(innerError);
          throw new Error(`Failed to parse JSON from code block: ${msg}`);
        }
      }
      const msg = e instanceof Error ? e.message : String(e);
      throw new Error(`Failed to parse JSON: ${msg}`);
    }
  }

  public extractTag(input: string, tagName: string): string | null {
    const escapedTag = tagName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`<${escapedTag}>([\\s\\S]*?)<\\/${escapedTag}>`, 'i');
    const match = input.match(regex);
    return match ? match[1] : null;
  }
}
