import { GoogleGenAI } from "@google/genai";

const apiKey = process.env.API_KEY || '';
let ai: GoogleGenAI | null = null;

if (apiKey) {
  ai = new GoogleGenAI({ apiKey });
}

export const generateMedicalResponse = async (prompt: string): Promise<string> => {
  if (!ai) {
    return "AI 助手未连接 (缺少 API Key)。请配置环境变量后重试。";
  }

  try {
    const response = await ai.models.generateContent({
      model: 'gemini-2.5-flash',
      contents: prompt,
      config: {
        systemInstruction: "你是一位 题酷 平台的智能医学助教。请用简洁、专业且易懂的语言回答用户的医学相关问题。如果用户询问购买或订阅，请引导他们去商城。请始终提醒用户：AI建议仅供参考，不可替代专业医疗诊断。",
        temperature: 0.7,
      }
    });

    return response.text || "抱歉，我现在无法回答。";
  } catch (error) {
    console.error("Gemini API Error:", error);
    return "抱歉，处理您的请求时出现错误。";
  }
};