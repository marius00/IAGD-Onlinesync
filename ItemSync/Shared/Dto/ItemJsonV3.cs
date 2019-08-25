namespace ItemSync.Shared.Dto {
    public class ItemJsonV3 : ItemJson {
        public string RemoteId { get; set; }
        public string RemotePartition { get; set; }
    }
}
