export const useCourseWithSyncedProgress = (course, notes = course?.notes || []) => {
  if (!course) return course

  const completed = notes.filter((note) => note?.status === "completed").length
  const total = notes.length
  const progress = total > 0 ? Math.round((completed / total) * 100) : 0

  return {
    ...course,
    notes,
    progress,
    status: total > 0 && completed === total ? "completed" : "in_progress",
  }
}
