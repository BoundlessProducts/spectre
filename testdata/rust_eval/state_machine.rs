// Finite state machine — enum-heavy, tuple variants, complex types.
use std::collections::HashMap;

pub enum State {
    Idle,
    Running { task_id: u32 },   // struct variant — data dropped
    Paused(u32),                // tuple variant — data dropped
    Done,
}

pub enum Event {
    Start,
    Pause,
    Resume,
    Stop,
}

pub struct Machine {
    pub current: State,
    pub transitions: HashMap<String, u32>,  // unsupported type
    pub history: Vec<String>,               // Vec<str>
    pub tick: u64,
}

impl Machine {
    pub fn new() -> Self {
        Machine {
            current: State::Idle,
            transitions: HashMap::new(),
            history: Vec::new(),
            tick: 0,
        }
    }

    pub fn process(&mut self, event: Event) {
        self.tick = self.tick.wrapping_add(1);
        self.history.push(format!("{}", self.tick));
        match event {
            Event::Start  => self.current = State::Running { task_id: 0 },
            Event::Pause  => self.current = State::Paused(0),
            Event::Resume => self.current = State::Running { task_id: 0 },
            Event::Stop   => self.current = State::Done,
        }
    }

    pub fn reset(&mut self) {
        self.current = State::Idle;
        self.tick = 0;
    }
}
