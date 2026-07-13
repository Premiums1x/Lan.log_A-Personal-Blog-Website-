-- Lancer identity content, applied exactly once even though the migration runner is idempotent.

CREATE TABLE IF NOT EXISTS migration_markers (
  key text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

UPDATE settings SET value = '{
  "brand":"Lancer",
  "footer_tag":"探索、失误、修正，然后继续进攻。这里记录每一回合留下的东西。",
  "since_year":2026,
  "commit_hash":"lancer26",
  "build_badge":"BREAK THE LINE / HOLD THE CLUTCH"
}'::jsonb, updated_at = now()
WHERE section_key = 'branding'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

UPDATE settings SET value = '{
  "eyebrow_cmd":"route --from today --to tomorrow",
  "title":"撕开防线，",
  "title_accent":"也守住残局",
  "title_tail":"。",
  "sub":"写下技术探索，也写下那些不顺利之后仍然决定继续向前的时刻。这里不是完成后的陈列柜，而是 Lancer 还在进行的这一回合。",
  "meta":[
    {"k":"focus","v":"frontend / go / agents"},
    {"k":"mode","v":"explore & build"},
    {"k":"mental","v":"clutch calm"}
  ],
  "corner":[]
}'::jsonb, updated_at = now()
WHERE section_key = 'hero'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

UPDATE settings SET value = '{
  "cells":[
    {"ic":"P01","title":"Paddock","desc":"F1 的速度、路线和对极限的判断。下一圈仍然敢于压上。"},
    {"ic":"P02","title":"Court / Pitch","desc":"Curry 与 Messi 把重复训练变成本能，在没有空间的位置创造空间。"},
    {"ic":"P03","title":"Server","desc":"Counter-Strike 的突破与残局：敢于进入，也要成为最后值得信任的人。"},
    {"ic":"P04","title":"Soundtrack","desc":"千禧华语、Phonk 与 EDM，为普通日子和想象中的高光时刻配乐。"}
  ]
}'::jsonb, updated_at = now()
WHERE section_key = 'stack'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

UPDATE settings SET value = '{
  "title":"有些路还没走明白，但我仍然相信更好的明天。",
  "title_accent":"更好的明天",
  "intro":[
    "我是 Lancer。这里不公开现实身份，只记录我愿意留下的思考、热爱和正在走的路线。",
    "我喜欢进攻带来的突破，也希望自己在真正的残局里足够冷静，让看到这一回合的人感到放心。"
  ],
  "meta":[
    {"k":"callsign","v":"Lancer"},
    {"k":"mode","v":"exploring"},
    {"k":"writes","v":"field notes"},
    {"k":"status","v":"round active"}
  ],
  "bio_yml":{"Name":"Lancer","Role":"explorer / builder","Stack":"frontend / go / agent","Writes":"field notes","Based":"private","Hosting":"self-hosted"},
  "uptime":"这一回合没结束",
  "body_md":"## 为什么是这个博客\n\n它首先写给我自己。不是为了把一段经历包装成漂亮结论，而是为了保留探索、失误、修正和继续向前的证据。\n\n## 进攻与残局\n\n我喜欢突破手撕开防线的瞬间，也喜欢残局里噪音逐渐消失、只剩判断和执行的时刻。写代码也很像这样：有时要大胆进入，有时要让自己慢下来，把问题一个个处理。\n\n## 更好的明天\n\n生活并不总按计划推进。偶尔不顺利时，我希望自己还能咬咬牙，把今天向前推一点。只要这一回合没有结束，就还有下一次选择。"
}'::jsonb, updated_at = now()
WHERE section_key = 'about'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

UPDATE settings SET value = '{
  "lines":[
    {"is_cmd":true,"f":"cat","args":"round.txt","c":"# current round"},
    {"arrow":"->","k":"building","v":"personal blog / field notes","is_string":true},
    {"arrow":"->","k":"learning","v":"frontend / go / agents","is_string":true},
    {"arrow":"->","k":"mental","v":"这一回合没结束","is_string":true},
    {"arrow":"->","k":"next","v":"把今天推进一点","is_string":true}
  ]
}'::jsonb, updated_at = now()
WHERE section_key = 'now'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

UPDATE settings SET value = '{
  "eyebrow_cmd":"season --all --newest-first",
  "title":"Season Log",
  "title_accent":"Season",
  "intro":"不是胜场统计，而是每一阶段真实留下的路线。回头看时，希望能看见自己怎样探索、失误、修正，然后继续向前。",
  "meta":[
    {"k":"mode","v":"season record"},
    {"k":"order","v":"newest first"},
    {"k":"status","v":"published rounds"}
  ]
}'::jsonb, updated_at = now()
WHERE section_key = 'archive'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

UPDATE settings SET value = '{
  "eyebrow_cmd":"loadout --show active",
  "title":"Loadout",
  "title_accent":"Loadout",
  "intro":"带进下一回合的工具、文档、故事、比赛和声音。它们不都直接关于代码，但都在塑造我的判断和热情。",
  "meta":[
    {"k":"type","v":"tools / stories / signals"},
    {"k":"update","v":"manual"},
    {"k":"owner","v":"Lancer"}
  ],
  "groups":[
    {"title":"Engineering","eyebrow":"active tools","desc":"帮助我把想法变成真实结果的工具和资料。","items":[
      {"title":"React Docs","desc":"回到组件、状态、Effect 与数据流的原点。","meta":"frontend","href":"https://react.dev/","status":"active","tags":["react","frontend"]},
      {"title":"Go by Example","desc":"用小例子训练后端语法、标准库和工程直觉。","meta":"backend","href":"https://gobyexample.com/","status":"active","tags":["go","backend"]}
    ]},
    {"title":"Competition","eyebrow":"mental models","desc":"从赛道、球场和服务器里学到的进攻与冷静。","items":[
      {"title":"Racecraft","desc":"路线、节奏、轮胎与风险选择。速度来自精确，也来自敢于行动。","meta":"F1","href":"","status":"watching","tags":["Verstappen","Leclerc"]},
      {"title":"Clutch Notes","desc":"突破先创造空间，残局再把噪音降到最低。","meta":"Counter-Strike","href":"","status":"training","tags":["NiKo","clutch"]}
    ]},
    {"title":"Soundtrack","eyebrow":"energy source","desc":"为长时间投入和想象中的高光时刻提供节奏。","items":[
      {"title":"Millennium Mandopop","desc":"周杰伦、邓紫棋，以及那些能把普通夜晚变成一段故事的歌。","meta":"playlist","href":"","status":"repeat","tags":["Mandopop","2000s"]},
      {"title":"Phonk / EDM","desc":"快节奏、低频和推进感，适合需要重新进入状态的时刻。","meta":"focus mix","href":"","status":"playing","tags":["phonk","edm"]}
    ]}
  ]
}'::jsonb, updated_at = now()
WHERE section_key = 'shelf'
  AND NOT EXISTS (SELECT 1 FROM migration_markers WHERE key = 'lancer_identity_v1');

INSERT INTO migration_markers (key)
VALUES ('lancer_identity_v1')
ON CONFLICT (key) DO NOTHING;