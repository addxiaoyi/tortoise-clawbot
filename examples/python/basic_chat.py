"""
Tortoise Python SDK - Quick Start Examples
"""

import asyncio
from tortoise import TortoiseClient


async def basic_chat():
    """Basic chat example"""
    async with TortoiseClient(base_url="http://localhost:18792") as client:
        # Create session
        session = await client.sessions.create(
            user_id="user@example.com",
            config={"model": "gpt-4o", "temperature": 0.7}
        )
        print(f"Session: {session.id}")

        # Send message
        response = await client.messages.send(
            session.id,
            content="Hello! How are you?"
        )
        print(f"Response: {response.content}")


async def streaming_chat():
    """Streaming response example"""
    async with TortoiseClient(base_url="http://localhost:18792") as client:
        session = await client.sessions.create()

        # Stream response
        print("Streaming response:\n")
        async for event in await client.messages.send(
            session.id,
            content="Count to 5",
            stream=True
        ):
            if event["type"] == "content_chunk":
                print(event["data"]["delta"], end="", flush=True)
        print()


async def tool_usage():
    """Tool usage example"""
    async with TortoiseClient(base_url="http://localhost:18792") as client:
        # List tools
        tools = await client.tools.list()
        print(f"Available tools: {[t.name for t in tools]}")

        # Invoke calculator
        result = await client.tools.invoke(
            "calculator",
            arguments={"expression": "100 / 5 + 10"}
        )
        print(f"Calculator result: {result.result}")


async def memory_example():
    """Memory management example"""
    async with TortoiseClient(base_url="http://localhost:18792") as client:
        # Store a memory
        stored = await client.memory.store(
            content="User prefers dark mode",
            type="preference",
            tags=["ui", "theme"],
            importance=0.8
        )
        print(f"Stored memory: {stored['id']}")

        # Search memories
        results = await client.memory.search(query="theme")
        print(f"Found {len(results)} memories")
        for mem in results:
            print(f"  - {mem.content} (similarity: {mem.importance})")


async def main():
    await basic_chat()
    await streaming_chat()
    await tool_usage()
    await memory_example()


if __name__ == "__main__":
    asyncio.run(main())
