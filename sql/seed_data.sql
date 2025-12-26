-- Заполнение таблицы users
-- Пароли будут автоматически захешированы через bcrypt при выполнении миграции
-- user1: user1123
-- user2: user2123  
-- admin: admin123
INSERT INTO users (id, login, password, is_moderator) VALUES
(1, 'user1', 'user1123', false),
(2, 'user2', 'user2123', false),
(3, 'admin', 'admin123', true);

-- Заполнение таблицы medications (лекарства)
-- adult_dose - доза для взрослого в мг (используется только для расчета детской дозы, не отображается в UI)
INSERT INTO medications (id, name, description, is_deleted, image_url, category, manufacturer, short_info, adult_dose) VALUES
(1, 'Нурофен детский', 'Нурофен применяется при головной боли, мигрени, зубной боли, повышенной температуре, невралгии, боли в ушах, мышечной и ревматической боли.', false, 'nurofen.png', 'Жаропонижающее', 'Reckitt', 'Ибупрофен 100 мг/5 мл', 200.00),
(2, 'Пенталгин', 'Пенталгин оказывает анальгезирующее и спазмолитическое действие.', false, 'pentalgin.png', 'Анальгетик', 'OTCpharm', 'Комбинированный анальгетик', 500.00),
(3, 'Гинкоум', 'Улучшает мозговое кровообращение, показан при снижении памяти и внимания.', false, 'ginkoum.png', 'Ноотроп', 'Evalar', 'Экстракт гинкго билоба', 80.00),
(4, 'Цитовир-3', 'Препарат с иммуномодулирующим действием для профилактики ОРВИ.', false, 'cytovir.png', 'Противовирусное', 'Петровакс', 'Иммуномодулятор', 200.00),
(5, 'Ринза', 'Комбинированное средство для снижения температуры и облегчения симптомов простуды.', false, 'rinza.png', 'От простуды', 'Unichem', 'При симптомах простуды', 500.00),
(6, 'Флуимуцил', 'Флуимуцил — это муколитический (секретолитический) препарат с прямым действием. Его основная задача — разжижать мокроту и облегчать ее выведение из дыхательных путей.', false, 'fluim.png', 'Муколитическое', 'ЗАМБОН', 'Муколитическое средство, разжижает мокроту', 600.00),
(7, 'Бронхо-мунал', 'Бронхо-Мунал — это иммуномодулирующий препарат бактериального происхождения. Он не является антибиотиком или противовирусным средством в прямом смысле.', false, 'bronhomun.png', 'Противовирусное', 'Sandoz', 'Бронхо-Мунал укрепляет иммунитет при простуде', 3.50);

-- Заполнение таблицы prescriptions (заявки)
INSERT INTO prescriptions (id, status, date_create, creator_id, date_formation, date_finish, moderator_id, patient_name, patient_gender, patient_weight, patient_age, calculation_result) VALUES
(1, 'черновик', '2024-09-20 10:15:30', 1, NULL, NULL, NULL, 'Алиса', 'Женский', 32.0, 8, NULL),
(2, 'завершён', '2024-09-15 09:00:00', 1, '2024-09-20 12:00:00', '2024-09-22 14:30:00', 2, 'Иван', 'Мужской', 25.0, 6, 'Итог: расчет дозы произведен врачом. Рекомендация: Нурофен 10 мг/кг'),
(3, 'сформирован', '2024-09-12 14:30:00', 1, '2024-09-19 09:45:00', NULL, 2, 'Мария', 'Женский', 18.0, 4, NULL),
(4, 'отклонён', '2024-09-17 16:05:00', 3, NULL, NULL, 2, 'Петр', 'Мужской', 30.0, 7, NULL);

-- Заполнение таблицы prescription_medications (связь заявок и лекарств)
INSERT INTO prescription_medications (id, prescription_id, medication_id, order_number, is_main, dosage_instruction) VALUES
(1, 1, 2, 1, false, 'по необходимости'),
(2, 1, 3, 2, false, 'курс 30 дней'),
(3, 2, 1, 1, true, '10 мг/кг, 3 раза в день'),
(4, 2, 4, 2, false, 'по инструкции'),
(5, 3, 5, 1, false, NULL),
(6, 4, 6, 1, false, NULL);

-- Синхронизация последовательностей ID после вставки данных с явными ID
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('medications_id_seq', (SELECT MAX(id) FROM medications));
SELECT setval('prescriptions_id_seq', (SELECT MAX(id) FROM prescriptions));
SELECT setval('prescription_medications_id_seq', (SELECT MAX(id) FROM prescription_medications));

