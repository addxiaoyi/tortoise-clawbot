import 'package:flutter/material.dart';

class MemoryPage extends StatefulWidget {
  const MemoryPage({super.key});

  @override
  State<MemoryPage> createState() => _MemoryPageState();
}

class _MemoryPageState extends State<MemoryPage> {
  final _searchController = TextEditingController();
  String _selectedFilter = 'All';

  final _filters = ['All', 'Facts', 'Preferences', 'Interests', 'Work'];

  final List<Map<String, dynamic>> _memories = [
    {'type': 'Facts', 'content': 'User prefers dark mode interface', 'importance': 4, 'time': '2h ago'},
    {'type': 'Preferences', 'content': 'Uses Python for data analysis tasks', 'importance': 5, 'time': '1d ago'},
    {'type': 'Interests', 'content': 'Interested in machine learning', 'importance': 3, 'time': '3d ago'},
    {'type': 'Work', 'content': 'Working on a new API project', 'importance': 4, 'time': '1w ago'},
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('Memory', style: TextStyle(fontWeight: FontWeight.w600)),
        elevation: 0,
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(24),
            child: TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: 'Search memories...',
                prefixIcon: const Icon(Icons.search),
                filled: true,
                fillColor: Colors.white,
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
              ),
            ),
          ),
          SizedBox(
            height: 40,
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 24),
              itemCount: _filters.length,
              itemBuilder: (context, index) {
                final filter = _filters[index];
                final isSelected = filter == _selectedFilter;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(filter),
                    selected: isSelected,
                    onSelected: (s) => setState(() => _selectedFilter = filter),
                    backgroundColor: Colors.white,
                    selectedColor: const Color(0xFF667EEA).withValues(alpha: 0.2),
                    checkmarkColor: const Color(0xFF667EEA),
                  ),
                );
              },
            ),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(horizontal: 24),
              itemCount: _memories.length,
              itemBuilder: (context, index) {
                final memory = _memories[index];
                return Container(
                  margin: const EdgeInsets.only(bottom: 12),
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 10)],
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                            decoration: BoxDecoration(
                              color: const Color(0xFF667EEA).withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(memory['type'], style: const TextStyle(fontSize: 12, color: Color(0xFF667EEA), fontWeight: FontWeight.w500)),
                          ),
                          const Spacer(),
                          Row(
                            children: List.generate(5, (i) => Icon(Icons.star, size: 14, color: i < memory['importance'] ? const Color(0xFFFBBF24) : Colors.grey.shade300)),
                          ),
                          const SizedBox(width: 8),
                          Text(memory['time'], style: const TextStyle(fontSize: 12, color: Color(0xFF94A3B8))),
                        ],
                      ),
                      const SizedBox(height: 12),
                      Text(memory['content'], style: const TextStyle(fontSize: 15, height: 1.5)),
                    ],
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
