import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { CoreAgentId } from '../lib/coreAgents';

interface CoreAgentState {
  selected: CoreAgentId | null;
  lastSelected: CoreAgentId | null;
  hasCompletedSelection: boolean;
  setSelectedCore: (core: CoreAgentId) => void;
  resetSelection: () => void;
  reenterSelection: () => void;
}

export const useCoreAgentStore = create<CoreAgentState>()(
  persist(
    (set) => ({
      selected: null,
      lastSelected: null,
      hasCompletedSelection: false,
      setSelectedCore: (core: CoreAgentId) =>
        set({
          selected: core,
          lastSelected: core,
          hasCompletedSelection: true,
        }),
      resetSelection: () =>
        set({
          selected: null,
          hasCompletedSelection: false,
        }),
      reenterSelection: () =>
        set({
          hasCompletedSelection: false,
        }),
    }),
    {
      name: 'core-agent-storage',
    }
  )
);
