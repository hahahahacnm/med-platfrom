
import { Product, Subject, WikiCategory } from './types';

// ==========================================
// 1. RAW QUIZ CONTENT
// ==========================================
const RAW_INTERNAL = `
### 泌尿系统疾病
1.关于慢性肾盂肾炎的描述，正确的是
A.肾小管功能受损较晚
B.多为大肠埃希菌感染
C.肾间质纤维化不明显
D.肾小球功能受损较早
E.由病毒感染引起
答案：B
解析：慢性肾盂肾炎常由细菌感染引起，其中大肠埃希菌最常见。

### 循环系统疾病
1.诊断冠心病最有价值的侵入性检查是
A.心腔内电生理检查
B.冠状动脉造影
C.心室造影
D.超声心动图
答案：B
解析：冠状动脉造影是诊断冠心病的“金标准”。

2.急性心肌梗死早期（24小时内）死亡的主要原因是
A.心力衰竭
B.心源性休克
C.心律失常
D.心脏破裂
答案：C
解析：心律失常是急性心肌梗死早期死亡的最主要原因，室颤最常见。
`;

const RAW_PATHOLOGY = `
### 细胞与组织的适应与损伤
1.细胞水肿发生的机制主要是
A.溶酶体膜受损
B.线粒体受损
C.高尔基体受损
D.内质网受损
答案：B
解析：细胞水肿主要由于线粒体受损，ATP生成减少，钠钾泵功能障碍。

2.下列哪项属于不可逆性损伤
A.细胞水肿
B.脂肪变性
C.核碎裂
D.玻璃样变性
答案：C
解析：核碎裂是细胞坏死的表现，属于不可逆损伤。

### 损伤的修复
1.肉芽组织的主要成分不包括
A.成纤维细胞
B.新生的毛细血管
C.炎性细胞
D.平滑肌细胞
答案：D
解析：肉芽组织由新生的毛细血管、成纤维细胞和炎性细胞组成。
`;

const RAW_SURGERY = `
### 外科休克
1.休克代偿期的临床表现是
A.血压稍低，脉快
B.血压正常或稍高，脉稍快，脉压压缩
C.血压稍低，脉快
D.血压正常，脉细速
答案：B
解析：休克代偿期，机体代偿导致外周血管收缩，血压可正常或稍高，脉压减小。

### 阑尾炎
1.急性阑尾炎最常见的病因是
A.细菌侵入
B.阑尾管腔阻塞
C.阑尾先天畸形
D.胃肠道功能紊乱
答案：B
解析：阑尾管腔阻塞是急性阑尾炎最常见的病因。
`;

// ==========================================
// 2. SUBJECTS (Quiz & Access Keys)
// Ids: 'internal', 'pathology', 'surgery'
// ==========================================
export const SUBJECTS: Subject[] = [
  {
    id: 'internal',
    title: '内科学',
    description: '呼吸、循环、消化、泌尿系统疾病诊疗',
    icon: '🫀',
    color: 'rose',
    rawContent: RAW_INTERNAL
  },
  {
    id: 'pathology',
    title: '病理学',
    description: '疾病的病因、发病机制与病理变化',
    icon: '🔬',
    color: 'violet',
    rawContent: RAW_PATHOLOGY
  },
  {
    id: 'surgery',
    title: '外科学',
    description: '创伤、感染、肿瘤的手术治疗',
    icon: '🔪',
    color: 'emerald',
    rawContent: RAW_SURGERY
  }
];

// ==========================================
// 3. WIKI DATA (Linked by ID)
// ==========================================
export const WIKI_CATEGORIES: WikiCategory[] = [
  {
    id: 'internal',
    title: '内科学知识库',
    description: '内科临床指南与最新研究',
    iconName: 'heart',
    color: 'rose',
    articles: [
      {
        id: 'htn-01',
        title: '高血压诊疗指南2024',
        excerpt: '最新高血压分级标准、药物治疗路径及生活方式干预。',
        content: `
高血压（Hypertension）是心血管疾病最重要的危险因素之一。它通常被称为"无声的杀手"，因为许多患者在早期没有任何症状。

## 1. 血压分类标准 (2024版)

根据最新的医学指南，成人血压分类如下表所示：

| 类别 | 收缩压 (mmHg) | 舒张压 (mmHg) | 处理建议 |
| :--- | :--- | :--- | :--- |
| **正常血压** | < 120 | < 80 | 每年监测 |
| **正常高值** | 120-139 | 80-89 | 生活方式干预 |
| **高血压 1 级** | 140-159 | 90-99 | 生活方式干预 + 药物治疗 |
| **高血压 2 级** | ≥ 160 | ≥ 100 | 立即启动药物治疗 |

## 2. 风险因素

高血压的成因复杂，通常分为可控与不可控因素：

*   **不可控因素**: 年龄、家族史、种族。
*   **可控因素**: 
    *   **肥胖**: 体重指数 (BMI) > 24。
    *   **高钠饮食**: 每日食盐摄入量 > 6g。
    *   **缺乏运动**: 每周少于 150 分钟中等强度运动。
    *   **精神压力**: 长期处于焦虑或高压状态。

> **临床警示**: 所有的抗高血压治疗都应建立在 **生活方式干预** 的基础上。仅靠药物而忽视饮食控制，往往难以达到理想的血压控制目标。

## 3. 药物治疗路径 (ABCD原则)

常用的一线降压药物包括以下四类：

| 缩写 | 药物类别 | 代表药物 | 适用人群 |
| :--- | :--- | :--- | :--- |
| **A** | ACEI / ARB | 普利类 / 沙坦类 | 糖尿病肾病、心力衰竭 |
| **B** | Beta-blocker | 洛尔类 | 冠心病、心率快者 |
| **C** | CCB | 地平类 | 老年高血压、收缩期高血压 |
| **D** | Diuretic | 噻嗪类 | 盐敏感性高血压、心衰 |

## 4. 特殊人群管理

### 老年高血压
老年人常表现为**单纯收缩期高血压**（即高压高，低压正常或偏低）。治疗时应注意从小剂量开始，避免体位性低血压。

### 妊娠期高血压
**禁用** ACEI 和 ARB 类药物，因为它们可能导致胎儿畸形。首选拉贝洛尔或甲基多巴。

---
*参考文献：2024 中国高血压防治指南, AHA Hypertension Guidelines.*
        `,
        author: '张主任',
        readTime: '8 min',
        date: '2024-03-15',
        tags: ['高血压', '慢病', '指南']
      }
    ]
  },
  {
    id: 'pathology',
    title: '病理学图谱',
    description: '高清病理切片解析',
    iconName: 'microscope',
    color: 'violet',
    articles: [
      {
        id: 'cell-injury',
        title: '细胞损伤与适应',
        excerpt: '可逆性损伤与不可逆性损伤的形态学区别。',
        content: '# 细胞损伤\n\n包括萎缩、肥大、增生、化生...',
        author: '李教授',
        readTime: '10 min',
        date: '2023-11-20',
        tags: ['基础医学']
      }
    ]
  },
  {
    id: 'surgery',
    title: '外科手术图解',
    description: '经典手术术式与解剖要点',
    iconName: 'scalpel',
    color: 'emerald',
    articles: []
  }
];

// ==========================================
// 4. PRODUCTS (Separated Quiz vs Wiki)
// AccessId convention: 'quiz_{id}' or 'wiki_{id}'
// ==========================================
export const PRODUCTS: Product[] = [
  // --- Internal Medicine ---
  {
    id: 'prod_internal_quiz',
    title: '内科学海量题库',
    description: '包含5000+道内科真题与模拟题，智能错题本功能。',
    price: 129,
    duration: '1年',
    imageUrl: 'https://images.unsplash.com/photo-1576091160399-112ba8d25d1d?auto=format&fit=crop&q=80&w=400',
    tags: ['题库', '内科', '刷题'],
    accessId: 'quiz_internal' 
  },
  {
    id: 'prod_internal_wiki',
    title: '内科临床知识库',
    description: '权威内科诊疗指南与专家视频解析，实时更新。',
    price: 99,
    duration: '1年',
    imageUrl: 'https://images.unsplash.com/photo-1532938911079-1b06ac7ceec7?auto=format&fit=crop&q=80&w=400',
    tags: ['知识库', '内科', '指南'],
    accessId: 'wiki_internal' 
  },

  // --- Pathology ---
  {
    id: 'prod_pathology_quiz',
    title: '病理学刷题通关包',
    description: '针对期末与考研的病理学专项训练，含名师解析。',
    price: 59,
    duration: '6个月',
    imageUrl: 'https://images.unsplash.com/photo-1579154204601-01588f351e67?auto=format&fit=crop&q=80&w=400',
    tags: ['题库', '基础', '病理'],
    accessId: 'quiz_pathology'
  },
  {
    id: 'prod_pathology_wiki',
    title: '病理学高清图谱库',
    description: '超过1000张高清病理切片图解，显微镜下的微观世界。',
    price: 69,
    duration: '1年',
    imageUrl: 'https://images.unsplash.com/photo-1530210124550-912dc1381cb8?auto=format&fit=crop&q=80&w=400',
    tags: ['知识库', '图谱', '病理'],
    accessId: 'wiki_pathology'
  },

  // --- Surgery ---
  {
    id: 'prod_surgery_quiz',
    title: '外科学专项练习',
    description: '涵盖普外、骨科、神经外科等分科试题。',
    price: 89,
    duration: '1年',
    imageUrl: 'https://images.unsplash.com/photo-1551076805-e1869033e561?auto=format&fit=crop&q=80&w=400',
    tags: ['题库', '外科', '真题'],
    accessId: 'quiz_surgery'
  },
  {
    id: 'prod_surgery_wiki',
    title: '外科学手术视频库',
    description: '经典手术术式演示与解剖要点全解析。',
    price: 109,
    duration: '1年',
    imageUrl: 'https://images.unsplash.com/photo-1516549655169-df83a092dd14?auto=format&fit=crop&q=80&w=400',
    tags: ['知识库', '视频', '外科'],
    accessId: 'wiki_surgery'
  }
];
