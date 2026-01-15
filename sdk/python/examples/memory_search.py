"""
Tortoise Python SDK - Memory Management Example
"""

import asyncio
from tortoise import TortoiseClient


async def main():
    async with TortoiseClient(
        gateway_url="http://localhost:8080",
        api_key="your-api-key"
    ) as client:
        # Add memories
        print("Adding memories...\n")
        
        await client.memory_add(
            content="User prefers to be called 'Alex'",
            memory_type="user_preference"
        )
        await client.memory_add(
            content="User is interested in Python programming",
            memory_type="interest"
        )
        await client.memory_add(
            content="User works as a software engineer at a startup",
            memory_type="work"
        )
        
        print("Added 3 memories\n")
        
        # Search memories
        print("Searching for 'programming'...\n")
        results = await client.memory_search(query="programming")
        
        for entry in results:
            print(f"- {entry['content']} ({entry['type']})")


if __name__ == "__main__":
    asyncio.run(main())
