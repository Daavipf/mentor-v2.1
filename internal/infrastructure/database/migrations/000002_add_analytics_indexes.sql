-- 000002_add_analytics_indexes.sql
-- Índices para otimização de consultas analíticas e relatórios de performance

CREATE INDEX IF NOT EXISTS idx_qoe_got_right ON questions_on_exams(exam_id, got_right);
CREATE INDEX IF NOT EXISTS idx_histories_finished ON histories(user_id, finished_at);
CREATE INDEX IF NOT EXISTS idx_qoe_updated_at ON questions_on_exams(updated_at);
CREATE INDEX IF NOT EXISTS idx_exams_user_date ON exams(user_id, date);

