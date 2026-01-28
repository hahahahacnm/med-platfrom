这是一个现代化的Web项目
前端使用 **React + TypeScript + Vite**
后端使用 **NestJS + TypeScript**


启动方式

cd backend
npm run start
npm run start:dev

/

npm run dev



BUG总览

1.注册、登录
2.



开发便签

账号: admin@med.edu
密码: admin123



管理员后台仪表盘累计刷题的饼状图，根据累计刷题的科目进行计算百分比，实现前后端，进行对接，开发，完善等




生产环境
PM2



管理员后台内容管理的题库管理，实现创建学科，然后可以批量上传章节题目文件，文件格式是.xlsx或者.csv,可批量上传多个文件，可上传文件夹，上传的文件根据文件名进行排序章节，比如‘1.绪论.xlsx’，‘2.无菌术.xlsx’

以下是xlsx表格文件排版例子，A1至P1

题号	科目路径	题型代码	题干	选项A	选项B	选项C	选项D	选项E	选项F	正确答案	解析	难易度	预测难度系数	大纲要求	认知层次
1	外科学 > 绪论 > 外科专业分科	A1	下列何种疾病属于外科疾病	骨折	慢性阻塞性肺疾病	糖尿病	慢性肾衰竭	病毒性肝炎		A		易	0.9	熟悉	解释
2	外科学 > 绪论 > 外科学发展简史	A1	我国最早对解剖学进行详细描述的著作是	《史记·扁鹊仓公列传》	《黄帝内经》	《难经》	《五脏图》	《医林改错》		C		易	0.9	熟悉	回忆


进行解析文件，保存数据库，实现前后端，进行对接，开发，完善等




管理员后台的商城运营，实现每个商品可设置多个优惠码，每个优惠码可单独设置，设置优惠码标记，减免金额，折扣，删除，使用次数等，售卖记录里可输入优惠码标记过滤，统计此优惠码标记售卖总金额，用户端商城购物车界面添加输入优惠码，实现前后端


实现管理员后台编辑商品的时长，可编辑天，月，年，用户端商城对接同步，用户订阅日期限制，实现可查看订阅日期，查看订阅到期时间




@zhifufm-main 是支付收款的NodeJs版本SDK，接口DEMO，请根据此SDK进行平台对接，前后端完善，能够正常使用





[Nest] 8836  - 2026/01/21 16:54:26     LOG [PaymentService] Creating order: f2936f02c93e4b8984785e5f54e41358, Amount: 1, Type: alipay
fetch {
  orderNo: 'f2936f02c93e4b8984785e5f54e41358',     
  amount: 1,
  payType: 'alipay',
  subject: 'test',
  returnType: 'json',
  returnUrl: 'http://med.owo.vin/',
  notifyUrl: 'http://medb.owo.vin/payment/notify', 
  merchantNum: '607807646955094016',
  sign: '9f77268179ed0e27fec3a3d25811174d'
} {
  success: true,
  msg: 'success',
  code: 200,
  timestamp: 1768985668378,
  data: {
    id: '608065746513313792',
    payUrl: 'https://alipayweixin.it88168.com/payali?orderNo=608065746513313792',
    extendParams: null
  }
}
[Nest] 8836  - 2026/01/21 16:54:28     LOG [PaymentService] Order created: {"success":true,"msg":"success","code":200,"timestamp":1768985668378,"data":{"id":"608065746513313792","payUrl":"https://alipayweixin.it88168.com/payali?orderNo=608065746513313792","extendParams":null}}
[Nest] 8836  - 2026/01/21 16:56:17     LOG [PaymentService] Creating order: 26ab096afa3f497483c304ebce71ef9b, Amount: 1, Type: alipay
fetch {
  orderNo: '26ab096afa3f497483c304ebce71ef9b',     
  amount: 1,
  payType: 'alipay',
  subject: 'test',
  returnType: 'json',
  returnUrl: 'http://med.owo.vin/',
  notifyUrl: 'http://medb.owo.vin/payment/notify', 
  merchantNum: '607807646955094016',
  sign: '008ac974df44944528b6ffd40e173297'
} {
  success: true,
  msg: 'success',
  code: 200,
  timestamp: 1768985778889,
  data: {
    id: '608066210063597568',
    payUrl: 'https://alipayweixin.it88168.com/payali?orderNo=608066210063597568',
    extendParams: null
  }
}
[Nest] 8836  - 2026/01/21 16:56:18     LOG [PaymentService] Order created: {"success":true,"msg":"success","code":200,"timestamp":1768985778889,"data":{"id":"608066210063597568","payUrl":"https://alipayweixin.it88168.com/payali?orderNo=608066210063597568","extendParams":null}}
[Nest] 8836  - 2026/01/21 16:57:58     LOG [PaymentService] Creating order: 66534fa88e9c40ceb54b9820456342c9, Amount: 1, Type: alipay
fetch {
  orderNo: '66534fa88e9c40ceb54b9820456342c9',     
  amount: 1,
  payType: 'alipay',
  subject: 'test',
  returnType: 'json',
  returnUrl: 'http://med.owo.vin/',
  notifyUrl: 'http://medb.owo.vin/payment/notify', 
  merchantNum: '607807646955094016',
  sign: '6ef4eb79cab3cd4e77ba11ec8cc840ca'
} {
  success: true,
  msg: 'success',
  code: 200,
  timestamp: 1768985879563,
  data: {
    id: '608066632266432512',
    payUrl: 'https://alipayweixin.it88168.com/payali?orderNo=608066632266432512',
    extendParams: null
  }
}
[Nest] 8836  - 2026/01/21 16:57:59     LOG [PaymentService] Order created: {"success":true,"msg":"success","code":200,"timestamp":1768985879563,"data":{"id":"608066632266432512","payUrl":"https://alipayweixin.it88168.com/payali?orderNo=608066632266432512","extendParams":null}}
