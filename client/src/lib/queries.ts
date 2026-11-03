import { useQuery } from "@tanstack/react-query";
import {
  getGlobalLogs,
  getGlobalDirectory,
  getPayloadStats,
  getDirInfo,
  getDirectoryList,
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
export function useDirectoryList() {
  return useQuery({
    queryKey: ["directory-list"],
    queryFn: getDirectoryList,
    refetchInterval: 12_000,
    staleTime: 5_000,
  });
}

export function useGlobalDirectory(path: string) {
  return useQuery({
    queryKey: ["global-directory", path],
    queryFn: () => getGlobalDirectory(path),
    enabled: Boolean(path),
    refetchInterval: 12_000,
    staleTime: 5_000,
  });
}

export function useDirInfo(path: string) {
  return useQuery({
    queryKey: ["directory-info", path],
    queryFn: () => getDirInfo(path),
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
