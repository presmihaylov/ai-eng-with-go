import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || '';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const notesApi = {
  getAll: () => api.get('/notes'),
  getById: (id) => api.get(`/notes/${id}`),
  create: (noteData) => api.post('/notes', noteData),
  update: (id, noteData) => api.put(`/notes/${id}`, noteData),
  delete: (id) => api.delete(`/notes/${id}`),
};

export const quizApi = {
  generate: (noteIDs, messages) => api.post('/quiz/generate', { note_ids: noteIDs, messages }),
  generateStream: (noteIDs, messages, onToken, onComplete, onError) => {
    return new Promise((resolve, reject) => {
      let accumulatedContent = '';
      
      // Send the request data via POST to initialize the stream
      fetch(`${API_BASE_URL}/quiz/generate/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ note_ids: noteIDs, messages }),
      }).then(response => {
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Set up EventSource for the streaming response
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        
        const readStream = () => {
          reader.read().then(({ done, value }) => {
            if (done) {
              onComplete(accumulatedContent);
              resolve(accumulatedContent);
              return;
            }
            
            const chunk = decoder.decode(value);
            const lines = chunk.split('\n');
            
            for (const line of lines) {
              accumulatedContent += line;
              onToken(line, accumulatedContent);
            }
            
            readStream();
          }).catch(error => {
            onError(error);
            reject(error);
          });
        };
        
        readStream();
      }).catch(error => {
        onError(error);
        reject(error);
      });
    });
  },
};

export const interactiveQuizApi = {
  configure: (messages) => api.post('/quiz/configure', { messages }),
  rank: (noteIDs, topics) => api.post('/quiz/rank', { note_ids: noteIDs, topics }),
  conduct: (noteIDs, topics, messages) => api.post('/quiz/conduct', { 
    note_ids: noteIDs, 
    topics: topics, 
    messages 
  })
};

export const quizV2Api = {
  configure: (messages) => api.post('/quiz/v2/configure', { messages }),
  createQuiz: (config) => api.post('/quizzes', { config }),
  conductQuiz: (quizId, messages) => api.post('/quiz/v2/conduct', { 
    quiz_id: quizId, 
    messages 
  }),
  getQuiz: (id) => api.get(`/quizzes/${id}`),
  getAllQuizzes: () => api.get('/quizzes'),
  updateQuiz: (id, updateData) => api.put(`/quizzes/${id}`, updateData),
  deleteQuiz: (id) => api.delete(`/quizzes/${id}`)
};

export const agentApi = {
  chat: (messages) => api.post('/agent/chat', { messages }),
};

export const healthApi = {
  check: () => api.get('/health'),
};
