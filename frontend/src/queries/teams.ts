import { useQuery } from '@tanstack/vue-query';
import { computed, type MaybeRef, toValue } from 'vue';
import { TeamService } from '@/services/team.service';
import { queryKeys } from './keys';

export function useTeams(orgId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.teams(toValue(orgId))),
        queryFn: () => TeamService.getTeams(toValue(orgId)),
        enabled: computed(() => !!toValue(orgId)),
    });
}

export function useTeamMembers(teamId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.teamMembers(toValue(teamId))),
        queryFn: () => TeamService.getTeamMembers(toValue(teamId)),
        enabled: computed(() => !!toValue(teamId)),
    });
}
