"""
Tortoise Python SDK - Plugin Management Example
"""

import asyncio
from tortoise import TortoiseClient


async def main():
    async with TortoiseClient(
        gateway_url="http://localhost:8080",
        api_key="your-api-key"
    ) as client:
        # List plugins
        plugins = await client.list_plugins()
        print("Available plugins:")
        for plugin in plugins:
            status = "enabled" if plugin.get("enabled") else "disabled"
            print(f"  - {plugin['name']} v{plugin['version']}: {status}")
        
        # Invoke a plugin
        print("\nInvoking 'web-search' plugin...")
        result = await client.invoke_plugin(
            plugin_id="web-search",
            action="search",
            params={"query": "Python async programming"}
        )
        print(f"Search results: {result}")


if __name__ == "__main__":
    asyncio.run(main())
