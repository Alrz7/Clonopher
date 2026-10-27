import { useQuery } from "@tanstack/react-query";
import {
  getGlobalLogs,
  getGlobalStage,
  getPayloadStats,
  getStageInfo,
  getStageList,
} from "./api";

export function usePayloadStats() {
  return useQuery({
    queryKey: ["payloads"],
    queryFn: getPayloadStats,
    refetchInterval: 4_000,
    refetchOnWindowFocus: true,
    staleTime: 1_000,
  });
}

export function useGLobalLogs() {
  return useQuery({
    queryKey: ["globalLogs"],
    queryFn: getPayloadStats,
    refetchInterval: 4_000,
    refetchOnWindowFocus: true,
    staleTime: 1_000,
  });
}
export function useStageList() {
  return useQuery({
    queryKey: ["stage-list"],
    queryFn: getStageList,
    refetchInterval: 12_000,
    staleTime: 5_000,
  });
}

export function useGlobalStage(path: string) {
  return useQuery({
    queryKey: ["global-stage", path],
    queryFn: () => getGlobalStage(path),
    enabled: Boolean(path),
    refetchInterval: 12_000,
    staleTime: 5_000,
  });
}

export function useStageInfo(path: string) {
  return useQuery({
    queryKey: ["stage-info", path],
    queryFn: () => getStageInfo(path),
    enabled: Boolean(path),
    staleTime: 10_000,
    refetchInterval: 15_000,
  });
}

export function useGlobalLogs(enabled = true) {
  return useQuery({
    queryKey: ["global-logs"],
    queryFn: getGlobalLogs,
    enabled,
    refetchInterval: enabled ? 2_000 : false,
    staleTime: 1_000,
  });
}
