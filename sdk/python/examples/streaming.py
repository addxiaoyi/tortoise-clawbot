"""
Tortoise Python SDK - 流式响应示例
演示如何处理流式AI响应
"""

import asyncio
from tortoise import Tortoise


async def streaming_chat():
    """流式聊天示例"""
    async with Tortoise(api_key="your-api-key") as client:
        # 创建会话
        session = client.create_session()
        
        # 发送消息并获取流式响应
        print("AI: ", end="", flush=True)
        
        async for chunk in session.chat_stream(
            "用Python写一个快速排序算法",
            model="gpt-4"
        ):
            print(chunk.content, end="", flush=True)
        
        print("\n")  # 换行


async def streaming_with_callback():
    """带回调的流式响应"""
    async with Tortoise(api_key="your-api-key") as client:
        session = client.create_session()
        
        chunks = []
        
        def on_chunk(chunk):
            chunks.append(chunk.content)
            print(chunk.content, end="", flush=True)
        
        def on_complete(response):
            print(f"\n\n完成! 总共 {len(chunks)} 个块")
        
        await session.chat_stream(
            "解释什么是装饰器模式",
            model="claude-3-sonnet",
            on_chunk=on_chunk,
            on_complete=on_complete
        )


async def streaming_pipeline():
    """流式处理管道"""
    async with Tortoise(api_key="your-api-key") as client:
        session = client.create_session()
        
        # 第一个问题
        print("问题1: 什么是Python的asyncio?\n")
        response1 = ""
        async for chunk in session.chat_stream("解释Python的asyncio"):
            response1 += chunk.content
            print(chunk.content, end="", flush=True)
        
        # 使用前一个答案作为上下文
        print("\n\n问题2 (基于上文):\n")
        async for chunk in session.chat_stream(
            f"基于这个解释：{response1[:100]}... 请给一个实际例子"
        ):
            print(chunk.content, end="", flush=True)


if __name__ == "__main__":
    print("=== 流式聊天示例 ===")
    asyncio.run(streaming_chat())
    
    print("\n=== 带回调的流式响应 ===")
    asyncio.run(streaming_with_callback())
