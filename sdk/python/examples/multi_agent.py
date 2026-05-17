"""
Tortoise Python SDK - 多代理示例
演示如何创建和协调多个代理
"""

import asyncio
from tortoise import Tortoise, Agent, Message


async def multi_agent_example():
    """多代理协调示例"""
    async with Tortoise(api_key="your-api-key") as client:
        # 创建协调器代理
        coordinator = Agent(
            name="coordinator",
            model="gpt-4",
            instructions="你是协调者，负责分解任务并分配给适当的代理。"
        )
        
        # 创建研究员代理
        researcher = Agent(
            name="researcher",
            model="gpt-4",
            instructions="你是一个研究员，负责搜索和分析信息。"
        )
        
        # 创建程序员代理
        coder = Agent(
            name="coder",
            model="gpt-4",
            instructions="你是一个程序员，负责编写高质量代码。"
        )
        
        # 注册代理
        client.register_agent(coordinator)
        client.register_agent(researcher)
        client.register_agent(coder)
        
        # 协调多个代理完成任务
        task = "研究并实现一个简单的Web服务器"
        
        # 1. 协调者分解任务
        plan = await coordinator.think(
            f"请将以下任务分解为子任务：{task}"
        )
        print(f"计划: {plan}")
        
        # 2. 研究员收集信息
        research = await researcher.search(
            query="Python asyncio web server best practices"
        )
        print(f"研究结果: {research}")
        
        # 3. 程序员编写代码
        code = await coder.generate(
            prompt=f"基于以下研究结果，生成Python Web服务器代码：{research}"
        )
        print(f"生成的代码: {code[:100]}...")
        
        return {
            "plan": plan,
            "research": research,
            "code": code
        }


async def agent_workflow():
    """代理工作流示例"""
    async with Tortoise(api_key="your-api-key") as client:
        # 创建工作流
        workflow = client.create_workflow(
            name="代码审查流程",
            agents=["researcher", "coder", "critic"]
        )
        
        # 添加任务
        workflow.add_task(
            "research",
            agent="researcher",
            prompt="研究Python代码审查的最佳实践"
        )
        
        workflow.add_task(
            "code",
            agent="coder", 
            prompt="编写一个代码审查工具"
        )
        workflow.depends_on("code", "research")
        
        workflow.add_task(
            "review",
            agent="critic",
            prompt="审查生成的代码并提供改进建议"
        )
        workflow.depends_on("review", "code")
        
        # 执行工作流
        result = await workflow.execute()
        
        return result


if __name__ == "__main__":
    # 运行示例
    result = asyncio.run(multi_agent_example())
    print(f"\n最终结果: {result}")
