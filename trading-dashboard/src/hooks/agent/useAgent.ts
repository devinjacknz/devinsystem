import { useState, useCallback, useEffect } from 'react';
import { useWebSocket } from '../websocket/useWebSocket';
import { Agent, AgentStrategy, TradingMode, CreateAgentRequest, UpdateAgentRequest } from '../../types/agent';
import { API_URL } from '../../utils/env';

export function useAgent() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { send, lastMessage } = useWebSocket(`${API_URL}/ws/agents`, {
    onMessage: (data) => {
      if (data.type === 'agent_update') {
        setAgents(prev => prev.map(agent => 
          agent.id === data.data.id ? { ...agent, ...data.data } : agent
        ));
      }
    }
  });

  const fetchAgents = useCallback(async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`${API_URL}/api/agents`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch agents');
      
      const data = await response.json();
      setAgents(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch agents');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createAgent = useCallback(async (request: CreateAgentRequest): Promise<Agent> => {
    try {
      const response = await fetch(`${API_URL}/api/agents`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        },
        body: JSON.stringify(request)
      });

      if (!response.ok) throw new Error('Failed to create agent');

      const agent = await response.json();
      setAgents(prev => [...prev, agent]);
      return agent;
    } catch (err) {
      throw err instanceof Error ? err : new Error('Failed to create agent');
    }
  }, []);

  const updateAgent = useCallback(async (request: UpdateAgentRequest): Promise<Agent> => {
    try {
      const response = await fetch(`${API_URL}/api/agents/${request.id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        },
        body: JSON.stringify(request)
      });

      if (!response.ok) throw new Error('Failed to update agent');

      const agent = await response.json();
      setAgents(prev => prev.map(a => a.id === agent.id ? agent : a));
      return agent;
    } catch (err) {
      throw err instanceof Error ? err : new Error('Failed to update agent');
    }
  }, []);

  const deleteAgent = useCallback(async (id: string): Promise<void> => {
    try {
      const response = await fetch(`${API_URL}/api/agents/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to delete agent');

      setAgents(prev => prev.filter(a => a.id !== id));
    } catch (err) {
      throw err instanceof Error ? err : new Error('Failed to delete agent');
    }
  }, []);

  const toggleAgent = useCallback(async (id: string, active: boolean): Promise<void> => {
    try {
      const response = await fetch(`${API_URL}/api/agents/${id}/${active ? 'start' : 'stop'}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      });

      if (!response.ok) throw new Error(`Failed to ${active ? 'start' : 'stop'} agent`);

      setAgents(prev => prev.map(a => 
        a.id === id ? { ...a, status: { ...a.status, status: active ? 'active' : 'paused' }} : a
      ));
    } catch (err) {
      throw err instanceof Error ? err : new Error(`Failed to ${active ? 'start' : 'stop'} agent`);
    }
  }, []);

  useEffect(() => {
    fetchAgents();
  }, [fetchAgents]);

  return {
    agents,
    isLoading,
    error,
    createAgent,
    updateAgent,
    deleteAgent,
    toggleAgent
  };
}
