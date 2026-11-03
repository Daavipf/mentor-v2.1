-- 000001_init_schema.sql
-- Migration reflecting schema.prisma for PostgreSQL

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context TEXT,
    alternatives_introduction TEXT,
    area VARCHAR(255) NOT NULL,
    discipline VARCHAR(255),
    year INT NOT NULL,
    language VARCHAR(100),
    index INT NOT NULL,
    title VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS question_topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    topic VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS question_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS alternatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text TEXT,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    image_path TEXT
);

CREATE TABLE IF NOT EXISTS exams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exam_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exams_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    area VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS questions_on_exams (
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    got_right BOOLEAN,
    selected_alternative_id UUID REFERENCES alternatives(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (question_id, exam_id)
);

CREATE TABLE IF NOT EXISTS histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    score INT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_questions_area ON questions(area);
CREATE INDEX IF NOT EXISTS idx_questions_discipline ON questions(discipline);
CREATE INDEX IF NOT EXISTS idx_question_topics_question_id ON question_topics(question_id);
CREATE INDEX IF NOT EXISTS idx_question_images_question_id ON question_images(question_id);
CREATE INDEX IF NOT EXISTS idx_alternatives_question_id ON alternatives(question_id);
CREATE INDEX IF NOT EXISTS idx_exams_user_id ON exams(user_id);
CREATE INDEX IF NOT EXISTS idx_exam_areas_exams_id ON exam_areas(exams_id);
CREATE INDEX IF NOT EXISTS idx_questions_on_exams_exam_id ON questions_on_exams(exam_id);
CREATE INDEX IF NOT EXISTS idx_histories_user_id ON histories(user_id);
CREATE INDEX IF NOT EXISTS idx_histories_exam_id ON histories(exam_id);
CREATE INDEX IF NOT EXISTS idx_qoe_got_right ON questions_on_exams(exam_id, got_right);
CREATE INDEX IF NOT EXISTS idx_histories_finished ON histories(user_id, finished_at);
CREATE INDEX IF NOT EXISTS idx_qoe_updated_at ON questions_on_exams(updated_at);
CREATE INDEX IF NOT EXISTS idx_exams_user_date ON exams(user_id, date);

