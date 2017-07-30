<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\ValueObjects;


class QueryOptions {

    /**
     * @var array
     */
    protected $params;

    /**
     * @var int
     */
    protected $limit = 100;

    /**
     * @var int
     */
    protected $offset = 0;

    /**
     * @var array
     */
    protected $cluster_name = [];


    /**
     * QueryOptions constructor.
     * @param $params
     */
    public function __construct($params) {
        $this->params = $params;

        if (isset($this->params['limit'])) {
            $this->limit = (int)$this->params['limit'];
        }

        if (isset($this->params['offset'])) {
            $this->offset = (int)$this->params['offset'];
        }

        if (isset($this->params['cluster_name']) && is_array($this->params['cluster_name'])) {
            foreach ($this->params['cluster_name'] as $clusterName) {
                if (strlen(trim($clusterName)) > 0 && $clusterName != 'false') {
                    $this->cluster_name[] = trim($clusterName);
                }
            }
        }

    }

    /**
     * @return int
     */
    public function getLimit() {
        return $this->limit;
    }

    /**
     * @return int
     */
    public function getOffset() {
        return $this->offset;
    }

    /**
     * @return array
     */
    public function getClusterName() {
        return $this->cluster_name;
    }

    public function getClusterNameSize() {
        return sizeof($this->cluster_name);
    }

}
