import React, { useEffect, useState } from 'react';
import { useTodoStore } from '../stores';
import type { TodoStatus } from '../types/index';

const statusColors: Record<TodoStatus, string> = {
  TODO: 'bg-gray-600',
  IN_PROGRESS: 'bg-blue-600',
  DONE: 'bg-green-600',
  ARCHIVED: 'bg-gray-500',
};

const statusLabels: Record<TodoStatus, string> = {
  TODO: 'To Do',
  IN_PROGRESS: 'In Progress',
  DONE: 'Done',
  ARCHIVED: 'Archived',
};

export const TodoView: React.FC = () => {
  const {
    todos,
    projects,
    isLoading,
    loadTodos,
    loadProjects,
    createTodo,
    moveTodo,
    deleteTodo
  } = useTodoStore();

  const [newTaskName, setNewTaskName] = useState('');
  const [selectedProject, setSelectedProject] = useState<string>('');

  useEffect(() => {
    loadTodos();
    loadProjects();
  }, [loadTodos, loadProjects]);

  const handleCreateTodo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTaskName.trim()) return;

    await createTodo({
      task_name: newTaskName,
      project: selectedProject || undefined,
    });
    setNewTaskName('');
  };

  const handleStatusChange = async (id: string, newStatus: TodoStatus) => {
    await moveTodo(id, newStatus);
  };

  if (isLoading && todos.length === 0) {
    return <div className="p-8 text-gray-400">Loading todos...</div>;
  }

  return (
    <div className="p-8">
      <div className="mb-6">
        <h2 className="text-2xl font-bold mb-4">Todos</h2>
        
        <form onSubmit={handleCreateTodo} className="flex gap-3">
          <input
            type="text"
            value={newTaskName}
            onChange={(e) => setNewTaskName(e.target.value)}
            placeholder="Add a new task..."
            className="flex-1 px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <select
            value={selectedProject}
            onChange={(e) => setSelectedProject(e.target.value)}
            className="px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
          >
            <option value="">No Project</option>
            {projects.map((project) => (
              <option key={project.name} value={project.name}>
                {project.name}
              </option>
            ))}
          </select>
          <button
            type="submit"
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
          >
            Add
          </button>
        </form>
      </div>

      {/* Projects summary */}
      {projects.length > 0 && (
        <div className="mb-6 flex gap-3 flex-wrap">
          {projects.map((project) => (
            <div
              key={project.name}
              className="px-3 py-1 bg-gray-800 rounded-full text-sm text-gray-300"
            >
              {project.name}: {project.todo_count}
            </div>
          ))}
        </div>
      )}

      {/* Todos list */}
      <div className="space-y-3">
        {todos.filter(t => t.status !== 'ARCHIVED').map((todo) => (
          <div
            key={todo.id}
            className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg border border-gray-700"
          >
            <select
              value={todo.status}
              onChange={(e) => handleStatusChange(todo.id, e.target.value as TodoStatus)}
              className={`px-2 py-1 rounded text-xs font-medium text-white ${statusColors[todo.status]}`}
            >
              {Object.entries(statusLabels).map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>

            <div className="flex-1">
              <p className={`font-medium ${todo.status === 'DONE' ? 'line-through text-gray-500' : 'text-white'}`}>
                {todo.task_name}
              </p>
              {todo.task_desc && (
                <p className="text-sm text-gray-400">{todo.task_desc}</p>
              )}
              {todo.project && (
                <span className="text-xs text-gray-500">{todo.project}</span>
              )}
            </div>

            {todo.labels.length > 0 && (
              <div className="flex gap-1">
                {todo.labels.map((label) => (
                  <span key={label} className="px-2 py-0.5 bg-gray-700 rounded text-xs text-gray-300">
                    {label}
                  </span>
                ))}
              </div>
            )}

            <button
              onClick={() => deleteTodo(todo.id)}
              className="text-red-400 hover:text-red-300 text-sm"
            >
              Delete
            </button>
          </div>
        ))}
      </div>

      {todos.filter(t => t.status !== 'ARCHIVED').length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <p>No todos yet. Add one to get started!</p>
        </div>
      )}
    </div>
  );
};
